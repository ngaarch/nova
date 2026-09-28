package topcmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFlagsParse(t *testing.T) {
	opts := ParseFlags([]string{"--sort=mem", "-n", "10", "--filter=systemd"})
	if opts.SortBy != "mem" {
		t.Errorf("expected SortBy mem, got %s", opts.SortBy)
	}
	if opts.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", opts.Limit)
	}
	if opts.Filter != "systemd" {
		t.Errorf("expected Filter systemd, got %s", opts.Filter)
	}
}

func TestCollectStats(t *testing.T) {
	opts := Options{
		SortBy: "cpu",
		Limit:  5,
	}

	stats, err := CollectStats(opts)
	if err != nil {
		t.Fatalf("CollectStats failed: %v", err)
	}

	if stats.MemTotal == 0 {
		t.Errorf("expected non-zero MemTotal")
	}
	if len(stats.Processes) == 0 {
		t.Errorf("expected at least 1 process returned")
	}
}

func TestFilterAndSortProcesses(t *testing.T) {
	procs := []ProcInfo{
		{PID: 10, Name: "alpha", CommandLine: "/bin/alpha", CPUPercent: 12.0, RSS: 1024, State: "S", User: "root"},
		{PID: 2, Name: "beta", CommandLine: "/usr/bin/beta", CPUPercent: 45.0, RSS: 8192, State: "R", User: "alice"},
		{PID: 30, Name: "gamma", CommandLine: "/opt/gamma", CPUPercent: 5.0, RSS: 4096, State: "S", User: "bob"},
	}

	stats := &SystemStats{}

	// Test sort by CPU
	optsCPU := Options{SortBy: "cpu", Limit: 10}
	sortedCPU := filterAndSortProcesses(procs, optsCPU, stats)
	if sortedCPU[0].PID != 2 {
		t.Errorf("expected PID 2 with highest CPU, got %d", sortedCPU[0].PID)
	}

	// Test sort by Mem
	optsMem := Options{SortBy: "mem", Limit: 10}
	sortedMem := filterAndSortProcesses(procs, optsMem, stats)
	if sortedMem[0].PID != 2 {
		t.Errorf("expected PID 2 with highest Mem, got %d", sortedMem[0].PID)
	}

	// Test sort by PID
	optsPID := Options{SortBy: "pid", Limit: 10}
	sortedPID := filterAndSortProcesses(procs, optsPID, stats)
	if sortedPID[0].PID != 2 || sortedPID[1].PID != 10 || sortedPID[2].PID != 30 {
		t.Errorf("PID order mismatch: got %d, %d, %d", sortedPID[0].PID, sortedPID[1].PID, sortedPID[2].PID)
	}

	// Test filter by substring
	optsFilter := Options{SortBy: "cpu", Filter: "gamma", Limit: 10}
	filtered := filterAndSortProcesses(procs, optsFilter, stats)
	if len(filtered) != 1 || filtered[0].Name != "gamma" {
		t.Errorf("expected 1 result with gamma, got %d", len(filtered))
	}
}

func TestRenderPlain(t *testing.T) {
	buf := &bytes.Buffer{}
	stats := &SystemStats{
		Processes: []ProcInfo{
			{PID: 123, User: "testuser", State: "R", CPUPercent: 4.5, MemPercent: 1.2, RSS: 2048, Threads: 4, CommandLine: "testapp"},
		},
	}
	opts := Options{}
	RenderPlain(buf, stats, opts)

	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("PID\tUSER\tSTATE")) {
		t.Errorf("expected plain header, got %q", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("123\ttestuser\tR")) {
		t.Errorf("expected process row, got %q", out)
	}
}

func TestRenderDashboard(t *testing.T) {
	buf := &bytes.Buffer{}
	stats := &SystemStats{
		LoadAvg:      [3]float64{0.5, 0.4, 0.3},
		Uptime:       "2d 3h",
		CPUUsage:     25.0,
		MemTotal:     16 * 1024 * 1024 * 1024,
		MemUsed:      4 * 1024 * 1024 * 1024,
		MemPercent:   25.0,
		TotalProcs:   100,
		RunningProcs: 2,
		SleepProcs:   98,
		Processes: []ProcInfo{
			{PID: 1, User: "root", State: "S", CPUPercent: 0.1, MemPercent: 0.2, RSS: 1024 * 1024, Threads: 1, CommandLine: "/init"},
		},
	}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, buf, buf, config.Config{}, caps, theme.Get("default"), output.ModeHuman, nil)
	opts := Options{}
	RenderDashboard(buf, stats, opts, ctx)

	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("System & Process Monitor")) {
		t.Errorf("expected title in dashboard, got: %s", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Load:")) {
		t.Errorf("expected load average, got: %s", out)
	}
}

func TestRunCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "top" {
		t.Errorf("expected command name top, got %s", cmd.Name)
	}

	err := cmd.Run(ctx, []string{"--limit=3", "--plain"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("PID")) {
		t.Errorf("expected PID in plain output, got: %s", stdout.String())
	}

	// Test JSON execution
	stdout.Reset()
	stderr.Reset()
	ctx = command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModeJSON, nil)
	errJSON := cmd.Run(ctx, []string{"--limit=3", "--json"})
	if errJSON != nil {
		t.Fatalf("expected nil error for JSON, got %v", errJSON)
	}

	var parsed SystemStats
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v, raw: %s", err, stdout.String())
	}
	if parsed.MemTotal == 0 {
		t.Errorf("expected non-zero MemTotal in JSON")
	}
}
