package netcmd

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFlagsParse(t *testing.T) {
	opts := ParseFlags([]string{"scan", "google.com", "-p", "80,443,8080", "-c", "5", "--plain"})
	if opts.Action != "scan" {
		t.Errorf("expected Action scan, got %s", opts.Action)
	}
	if opts.Target != "google.com" {
		t.Errorf("expected Target google.com, got %s", opts.Target)
	}
	if len(opts.Ports) != 3 || opts.Ports[0] != 80 || opts.Ports[1] != 443 || opts.Ports[2] != 8080 {
		t.Errorf("expected ports [80, 443, 8080], got %v", opts.Ports)
	}
	if opts.Count != 5 {
		t.Errorf("expected Count 5, got %d", opts.Count)
	}
	if !opts.Plain {
		t.Errorf("expected Plain true")
	}
}

func TestExecutePingWithLocalServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start local test server: %v", err)
	}
	defer ln.Close()

	// Accept connections in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().String()
	opts := Options{
		Target:   addr,
		Count:    3,
		Interval: 10 * time.Millisecond,
		Timeout:  1 * time.Second,
	}

	res := ExecutePing(opts)
	if res.Transmitted != 3 {
		t.Errorf("expected 3 transmitted, got %d", res.Transmitted)
	}
	if res.Received != 3 {
		t.Errorf("expected 3 received, got %d", res.Received)
	}
	if res.LossPercent != 0.0 {
		t.Errorf("expected 0%% loss, got %.1f%%", res.LossPercent)
	}
	if res.AvgLatency <= 0 {
		t.Errorf("expected positive avg latency")
	}
}

func TestExecutePortScan(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start local test server: %v", err)
	}
	defer ln.Close()

	openPort := ln.Addr().(*net.TCPAddr).Port
	closedPort := openPort + 1
	if closedPort > 65535 {
		closedPort = openPort - 1
	}

	res := ExecutePortScan("127.0.0.1", []int{openPort, closedPort}, 500*time.Millisecond)
	if res.TotalScanned != 2 {
		t.Errorf("expected 2 scanned, got %d", res.TotalScanned)
	}

	foundOpen := false
	for _, p := range res.Ports {
		if p.Port == openPort && p.State == "open" {
			foundOpen = true
			break
		}
	}
	if !foundOpen {
		t.Errorf("expected port %d to be open", openPort)
	}
}

func TestRenderPingDashboard(t *testing.T) {
	buf := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, buf, buf, config.Config{}, caps, theme.Get("default"), output.ModeHuman, nil)

	res := &PingResult{
		Target:      "127.0.0.1",
		Address:     "127.0.0.1:80",
		Transmitted: 2,
		Received:    2,
		Attempts: []PingAttempt{
			{Seq: 1, Latency: 10 * time.Millisecond, Success: true},
			{Seq: 2, Latency: 15 * time.Millisecond, Success: true},
		},
		MinLatency: 10 * time.Millisecond,
		AvgLatency: 12500 * time.Microsecond,
		MaxLatency: 15 * time.Millisecond,
		Jitter:     2500 * time.Microsecond,
	}

	RenderPingDashboard(buf, res, ctx)
	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("TCP Ping Telemetry")) {
		t.Errorf("expected dashboard title, got: %s", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Packets: 2 sent, 2 received")) {
		t.Errorf("expected packet summary, got: %s", out)
	}
}

func TestRunCommand(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start local test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "net" {
		t.Errorf("expected name net, got %s", cmd.Name)
	}

	addr := ln.Addr().String()
	err = cmd.Run(ctx, []string{"ping", addr, "-c", "2", "--plain"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("SEQ\tSTATUS\tLATENCY_US")) {
		t.Errorf("expected plain ping header, got: %s", stdout.String())
	}

	// Test JSON mode
	stdout.Reset()
	ctx = command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModeJSON, nil)
	err = cmd.Run(ctx, []string{"ping", addr, "-c", "2", "--json"})
	if err != nil {
		t.Fatalf("expected nil error for JSON, got %v", err)
	}

	var parsed PingResult
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, raw: %s", err, stdout.String())
	}
	if parsed.Transmitted != 2 {
		t.Errorf("expected 2 transmitted in JSON, got %d", parsed.Transmitted)
	}
}
