package certcmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(stdin string, stdout, stderr *bytes.Buffer) *command.Context {
	cfg := config.Config{
		Theme:     "default",
		ColorMode: "never",
	}
	caps := terminal.Capabilities{
		Width:        80,
		Height:       24,
		ColorProfile: terminal.ColorNone,
		IsTTY:        false,
	}
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(
		strings.NewReader(stdin),
		stdout,
		stderr,
		cfg,
		caps,
		th,
		output.ModePlain,
		logger,
	)
}

func generateTestCertificate(t *testing.T, path string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(123456789),
		Subject: pkix.Name{
			CommonName:   "nova.test.local",
			Organization: []string{"Nova Test Org"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(90 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"nova.test.local", "api.nova.test.local"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	if err := os.WriteFile(path, certPEM, 0644); err != nil {
		t.Fatal(err)
	}
}

func generateTestJWT(t *testing.T, sub string, exp time.Time) string {
	t.Helper()
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	payload := map[string]any{
		"sub":  sub,
		"iss":  "nova.auth",
		"exp":  exp.Unix(),
		"iat":  exp.Add(-1 * time.Hour).Unix(),
		"role": "admin",
	}

	hJSON, _ := json.Marshal(header)
	pJSON, _ := json.Marshal(payload)

	hB64 := base64.RawURLEncoding.EncodeToString(hJSON)
	pB64 := base64.RawURLEncoding.EncodeToString(pJSON)
	sigB64 := base64.RawURLEncoding.EncodeToString([]byte("dummy-signature-bytes"))

	return hB64 + "." + pB64 + "." + sigB64
}

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"inspect", "example.com:443", "--timeout=10s"})
	if opts.Mode != ModeTLS || opts.Target != "example.com:443" || opts.Timeout != 10*time.Second {
		t.Errorf("unexpected TLS opts: %+v", opts)
	}

	jwtStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgN"
	optsJWT := ParseFlags([]string{jwtStr})
	if optsJWT.Mode != ModeJWT || optsJWT.Target != jwtStr {
		t.Errorf("unexpected JWT auto-detect opts: %+v", optsJWT)
	}
}

func TestDecodeJWT(t *testing.T) {
	futureExp := time.Now().Add(2 * time.Hour)
	token := generateTestJWT(t, "user-nova-42", futureExp)

	info, err := DecodeJWT(token)
	if err != nil {
		t.Fatalf("DecodeJWT failed: %v", err)
	}

	if info.Subject != "user-nova-42" {
		t.Errorf("expected subject user-nova-42, got %q", info.Subject)
	}
	if info.Algorithm != "HS256" || info.Type != "JWT" {
		t.Errorf("unexpected alg/type: %s / %s", info.Algorithm, info.Type)
	}
	if info.IsExpired {
		t.Errorf("expected token not expired")
	}
	if !strings.Contains(info.TimeStatus, "Valid for") {
		t.Errorf("expected TimeStatus to contain 'Valid for', got: %s", info.TimeStatus)
	}
}

func TestInspectLocalCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "server.crt")
	generateTestCertificate(t, certPath)

	res, err := InspectLocalFile(certPath)
	if err != nil {
		t.Fatalf("InspectLocalFile failed: %v", err)
	}

	if len(res.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(res.Certificates))
	}

	cert := res.Certificates[0]
	if cert.CommonName != "nova.test.local" {
		t.Errorf("expected CN nova.test.local, got %q", cert.CommonName)
	}
	if cert.Status != "Valid" {
		t.Errorf("expected Status Valid, got %q", cert.Status)
	}
	if len(cert.SANs) != 3 { // 2 DNS + 1 IP
		t.Errorf("expected 3 SANs, got %d", len(cert.SANs))
	}
}

func TestRunCommand_LocalCertAndJWT(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "server.crt")
	generateTestCertificate(t, certPath)

	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr)

	// 1. Inspect local cert in plain mode
	err := Run(ctx, []string{"inspect", certPath, "--plain"})
	if err != nil {
		t.Fatalf("Run failed for cert inspect: %v", err)
	}
	if !strings.Contains(stdout.String(), "nova.test.local") || !strings.Contains(stdout.String(), "Valid") {
		t.Errorf("expected nova.test.local and Valid in plain cert output: %s", stdout.String())
	}

	// 2. Inspect local cert in JSON mode
	stdout.Reset()
	stderr.Reset()
	err = Run(ctx, []string{certPath, "--json"})
	if err != nil {
		t.Fatalf("Run failed for cert json: %v", err)
	}
	var res CertResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode cert JSON: %v; raw: %s", err, stdout.String())
	}
	if len(res.Certificates) == 0 || res.Certificates[0].CommonName != "nova.test.local" {
		t.Errorf("unexpected decoded cert JSON: %+v", res)
	}

	// 3. Inspect JWT in plain mode
	stdout.Reset()
	stderr.Reset()
	token := generateTestJWT(t, "admin-agent", time.Now().Add(1*time.Hour))
	err = Run(ctx, []string{"jwt", token, "--plain"})
	if err != nil {
		t.Fatalf("Run failed for JWT: %v", err)
	}
	if !strings.Contains(stdout.String(), "admin-agent") || !strings.Contains(stdout.String(), "HS256") {
		t.Errorf("expected admin-agent and HS256 in JWT output: %s", stdout.String())
	}
}
