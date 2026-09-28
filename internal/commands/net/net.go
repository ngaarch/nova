package netcmd

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// PingAttempt represents one single probe.
type PingAttempt struct {
	Seq     int           `json:"seq"`
	Latency time.Duration `json:"latency"`
	Success bool          `json:"success"`
	Error   string        `json:"error,omitempty"`
}

// PingResult aggregates ping telemetry.
type PingResult struct {
	Target      string        `json:"target"`
	Address     string        `json:"address"`
	Transmitted int           `json:"transmitted"`
	Received    int           `json:"received"`
	LossPercent float64       `json:"loss_percent"`
	MinLatency  time.Duration `json:"min_latency"`
	AvgLatency  time.Duration `json:"avg_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	Jitter      time.Duration `json:"jitter"`
	Attempts    []PingAttempt `json:"attempts"`
}

// PortItem represents one probed port.
type PortItem struct {
	Port    int           `json:"port"`
	Service string        `json:"service"`
	State   string        `json:"state"` // "open", "closed", "timeout"
	Latency time.Duration `json:"latency"`
}

// PortScanResult holds scan findings.
type PortScanResult struct {
	Target       string        `json:"target"`
	TotalScanned int           `json:"total_scanned"`
	OpenCount    int           `json:"open_count"`
	ClosedCount  int           `json:"closed_count"`
	Duration     time.Duration `json:"duration"`
	Ports        []PortItem    `json:"ports"`
}

// DNSResult contains resolved records for a domain.
type DNSResult struct {
	Domain   string        `json:"domain"`
	A        []string      `json:"a_records"`
	AAAA     []string      `json:"aaaa_records"`
	CNAME    string        `json:"cname,omitempty"`
	MX       []string      `json:"mx_records,omitempty"`
	TXT      []string      `json:"txt_records,omitempty"`
	Duration time.Duration `json:"duration"`
}

// Command returns the registered Command instance for net.
func Command() *command.Command {
	return &command.Command{
		Name:        "net",
		Aliases:     []string{"ping", "latency", "portscan"},
		Summary:     "Network latency telemetry, high-precision TCP ping, and port scanner",
		Usage:       "nova net [ping|scan|dns] <target> [flags]",
		Description: "Perform socket latency diagnostics, concurrent port scanning, and DNS inspection.",
		Phase:       22,
		Run:         Run,
	}
}

// Run executes the net command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	switch opts.Action {
	case "scan":
		return runScan(ctx, opts)
	case "dns":
		return runDNS(ctx, opts)
	case "ping":
		fallthrough
	default:
		return runPing(ctx, opts)
	}
}

func runPing(ctx *command.Context, opts Options) error {
	res := ExecutePing(opts)

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderPingPlain(ctx.Stdout, res)
	} else {
		RenderPingDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

func runScan(ctx *command.Context, opts Options) error {
	res := ExecutePortScan(opts.Target, opts.Ports, opts.Timeout)

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderScanPlain(ctx.Stdout, res)
	} else {
		RenderScanDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

func runDNS(ctx *command.Context, opts Options) error {
	res, err := ExecuteDNS(opts.Target)
	if err != nil {
		return fmt.Errorf("dns query %s: %w", opts.Target, err)
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderDNSPlain(ctx.Stdout, res)
	} else {
		RenderDNSDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

// ExecutePing probes the target with TCP connections.
func ExecutePing(opts Options) *PingResult {
	addr := opts.Target
	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}

	res := &PingResult{
		Target:      opts.Target,
		Address:     addr,
		Transmitted: opts.Count,
	}

	var latencies []time.Duration

	for seq := 1; seq <= opts.Count; seq++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, opts.Timeout)
		dur := time.Since(start)

		att := PingAttempt{
			Seq:     seq,
			Latency: dur,
		}

		if err == nil {
			_ = conn.Close()
			att.Success = true
			res.Received++
			latencies = append(latencies, dur)
		} else {
			att.Success = false
			att.Error = err.Error()
		}

		res.Attempts = append(res.Attempts, att)

		if seq < opts.Count && opts.Interval > 0 {
			time.Sleep(opts.Interval)
		}
	}

	if res.Transmitted > 0 {
		res.LossPercent = float64(res.Transmitted-res.Received) / float64(res.Transmitted) * 100.0
	}

	if len(latencies) > 0 {
		minL := latencies[0]
		maxL := latencies[0]
		var total time.Duration

		for _, l := range latencies {
			if l < minL {
				minL = l
			}
			if l > maxL {
				maxL = l
			}
			total += l
		}

		avgL := total / time.Duration(len(latencies))
		res.MinLatency = minL
		res.MaxLatency = maxL
		res.AvgLatency = avgL

		// Calculate jitter (mean absolute deviation from avg)
		var devSum float64
		for _, l := range latencies {
			devSum += math.Abs(float64(l - avgL))
		}
		res.Jitter = time.Duration(devSum / float64(len(latencies)))
	}

	return res
}

// KnownPortServices returns standard service names for common ports.
var KnownPortServices = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	465:   "SMTPS",
	587:   "Submission",
	993:   "IMAPS",
	995:   "POP3S",
	3000:  "Node/Dev",
	3306:  "MySQL",
	5432:  "PostgreSQL",
	6379:  "Redis",
	8000:  "HTTP-Alt",
	8080:  "HTTP-Proxy",
	8443:  "HTTPS-Alt",
	27017: "MongoDB",
}

// ExecutePortScan scans specified ports concurrently.
func ExecutePortScan(target string, ports []int, timeout time.Duration) *PortScanResult {
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}

	start := time.Now()
	res := &PortScanResult{
		Target:       target,
		TotalScanned: len(ports),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 50) // concurrency limit

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			addr := fmt.Sprintf("%s:%d", target, p)
			connStart := time.Now()
			conn, err := net.DialTimeout("tcp", addr, timeout)
			lat := time.Since(connStart)

			item := PortItem{
				Port:    p,
				Service: KnownPortServices[p],
				Latency: lat,
			}
			if item.Service == "" {
				item.Service = "Custom"
			}

			if err == nil {
				_ = conn.Close()
				item.State = "open"
			} else {
				if strings.Contains(err.Error(), "timeout") {
					item.State = "timeout"
				} else {
					item.State = "closed"
				}
			}

			mu.Lock()
			res.Ports = append(res.Ports, item)
			if item.State == "open" {
				res.OpenCount++
			} else {
				res.ClosedCount++
			}
			mu.Unlock()
		}(port)
	}

	wg.Wait()
	res.Duration = time.Since(start)

	// Sort ports ascending
	sort.Slice(res.Ports, func(i, j int) bool {
		return res.Ports[i].Port < res.Ports[j].Port
	})

	return res
}

// ExecuteDNS queries DNS records for domain.
func ExecuteDNS(domain string) (*DNSResult, error) {
	start := time.Now()
	res := &DNSResult{
		Domain: domain,
	}

	// Lookup IPs
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := net.DefaultResolver

	ips, err := resolver.LookupIP(ctx, "ip", domain)
	if err == nil {
		for _, ip := range ips {
			if ip.To4() != nil {
				res.A = append(res.A, ip.String())
			} else {
				res.AAAA = append(res.AAAA, ip.String())
			}
		}
	}

	if cname, err := resolver.LookupCNAME(ctx, domain); err == nil && cname != domain+"." && cname != domain {
		res.CNAME = strings.TrimSuffix(cname, ".")
	}

	if mxs, err := resolver.LookupMX(ctx, domain); err == nil {
		for _, mx := range mxs {
			res.MX = append(res.MX, fmt.Sprintf("%s (pri %d)", strings.TrimSuffix(mx.Host, "."), mx.Pref))
		}
	}

	if txts, err := resolver.LookupTXT(ctx, domain); err == nil {
		res.TXT = append(res.TXT, txts...)
	}

	res.Duration = time.Since(start)
	return res, nil
}
