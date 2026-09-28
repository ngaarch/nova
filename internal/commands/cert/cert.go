package certcmd

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// CertInfo represents decoded X.509 certificate metadata.
type CertInfo struct {
	CommonName         string    `json:"common_name"`
	Subject            string    `json:"subject"`
	Issuer             string    `json:"issuer"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	DaysRemaining      int       `json:"days_remaining"`
	Status             string    `json:"status"` // "Valid", "Expiring Soon", "Expired"
	SANs               []string  `json:"sans,omitempty"`
	KeyAlgorithm       string    `json:"key_algorithm"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	SerialNumber       string    `json:"serial_number"`
	IsCA               bool      `json:"is_ca"`
}

// CertResult holds the full inspection payload for TLS host or local cert.
type CertResult struct {
	Target       string     `json:"target"`
	IsRemote     bool       `json:"is_remote"`
	TLSVersion   string     `json:"tls_version,omitempty"`
	CipherSuite  string     `json:"cipher_suite,omitempty"`
	ServerName   string     `json:"server_name,omitempty"`
	Certificates []CertInfo `json:"certificates"`
}

// Command returns the registered Command instance for cert.
func Command() *command.Command {
	return &command.Command{
		Name:        "cert",
		Aliases:     []string{"ssl", "tls", "jwt"},
		Summary:     "SSL/TLS certificate validator and JWT token decoder",
		Usage:       "nova cert [inspect <host|file> | jwt <token>] [flags]",
		Description: "Audit remote TLS certificates, verify local PEM/CRT files, and decode JWT tokens.",
		Phase:       29,
		Run:         Run,
	}
}

// Run executes the cert command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	if strings.TrimSpace(opts.Target) == "" {
		return fmt.Errorf("target required: e.g. nova cert inspect example.com:443, or nova cert jwt <token>")
	}

	if opts.Mode == ModeJWT {
		jwtInfo, err := DecodeJWT(opts.Target)
		if err != nil {
			return fmt.Errorf("jwt: %w", err)
		}

		if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
			return ctx.Printer.PrintJSON(jwtInfo)
		}

		if ctx.Printer.Mode == output.ModePlain || opts.Plain {
			RenderJWTPlain(ctx.Stdout, jwtInfo)
		} else {
			RenderJWTDashboard(ctx.Stdout, jwtInfo, ctx)
		}
		return nil
	}

	// TLS / Certificate Mode
	var res *CertResult
	if fi, err := os.Stat(opts.Target); err == nil && !fi.IsDir() {
		res, err = InspectLocalFile(opts.Target)
		if err != nil {
			return fmt.Errorf("inspect cert file: %w", err)
		}
	} else {
		var err error
		res, err = InspectRemote(opts.Target, opts.Timeout, opts.Insecure)
		if err != nil {
			return fmt.Errorf("inspect tls %q: %w", opts.Target, err)
		}
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderCertPlain(ctx.Stdout, res)
	} else {
		RenderCertDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

// InspectRemote connects to target host via TLS and extracts certificate chain.
func InspectRemote(target string, timeout time.Duration, insecure bool) (*CertResult, error) {
	host := target
	port := "443"

	if strings.Contains(target, ":") {
		h, p, err := net.SplitHostPort(target)
		if err == nil {
			host = h
			port = p
		}
	}
	addr := net.JoinHostPort(host, port)

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: insecure,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("tls connect to %s: %w", addr, err)
	}
	defer conn.Close()

	cs := conn.ConnectionState()
	res := &CertResult{
		Target:      addr,
		IsRemote:    true,
		TLSVersion:  tlsVersionString(cs.Version),
		CipherSuite: tls.CipherSuiteName(cs.CipherSuite),
		ServerName:  cs.ServerName,
	}

	for _, cert := range cs.PeerCertificates {
		res.Certificates = append(res.Certificates, convertCert(cert))
	}

	return res, nil
}

// InspectLocalFile reads and decodes X.509 certificates from a local PEM/CRT file.
func InspectLocalFile(path string) (*CertResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	res := &CertResult{
		Target:   path,
		IsRemote: false,
	}

	// Try decoding PEM blocks
	rest := data
	var certs []*x509.Certificate
	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			c, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				certs = append(certs, c)
			}
		}
	}

	// Fallback to DER encoding if no PEM blocks found
	if len(certs) == 0 {
		c, err := x509.ParseCertificate(data)
		if err != nil {
			return nil, fmt.Errorf("file contains no valid PEM or DER X.509 certificates")
		}
		certs = append(certs, c)
	}

	for _, c := range certs {
		res.Certificates = append(res.Certificates, convertCert(c))
	}

	return res, nil
}

func convertCert(c *x509.Certificate) CertInfo {
	now := time.Now()
	days := int(c.NotAfter.Sub(now).Hours() / 24)

	status := "Valid"
	if now.After(c.NotAfter) {
		status = "Expired"
	} else if days < 30 {
		status = "Expiring Soon"
	}

	var sans []string
	sans = append(sans, c.DNSNames...)
	for _, ip := range c.IPAddresses {
		sans = append(sans, ip.String())
	}

	keyAlgo := c.PublicKeyAlgorithm.String()
	switch pub := c.PublicKey.(type) {
	case *rsa.PublicKey:
		keyAlgo = fmt.Sprintf("RSA (%d bits)", pub.N.BitLen())
	case *ecdsa.PublicKey:
		keyAlgo = fmt.Sprintf("ECDSA (%s)", pub.Curve.Params().Name)
	}

	cn := c.Subject.CommonName
	if cn == "" && len(c.DNSNames) > 0 {
		cn = c.DNSNames[0]
	}
	if cn == "" {
		cn = "(no CN)"
	}

	return CertInfo{
		CommonName:         cn,
		Subject:            c.Subject.String(),
		Issuer:             c.Issuer.String(),
		NotBefore:          c.NotBefore,
		NotAfter:           c.NotAfter,
		DaysRemaining:      days,
		Status:             status,
		SANs:               sans,
		KeyAlgorithm:       keyAlgo,
		SignatureAlgorithm: c.SignatureAlgorithm.String(),
		SerialNumber:       formatSerial(c.SerialNumber.Bytes()),
		IsCA:               c.IsCA,
	}
}

func formatSerial(bytes []byte) string {
	var parts []string
	for _, b := range bytes {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, ":")
}

func tlsVersionString(ver uint16) string {
	switch ver {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return fmt.Sprintf("0x%04X", ver)
	}
}
