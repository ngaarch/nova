package bench

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderJSON outputs benchmark metrics in formatted JSON.
func RenderJSON(ctx *command.Context, res Results) error {
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

// RenderPlain prints metrics in parseable plain key=value format.
func RenderPlain(ctx *command.Context, res Results) {
	ctx.Printer.Printf("seq_write_mb_s=%.2f\n", res.SeqWriteMBs)
	ctx.Printer.Printf("seq_read_mb_s=%.2f\n", res.SeqReadMBs)
	ctx.Printer.Printf("stat_ops_per_sec=%.2f\n", res.StatOpsPerSec)
	ctx.Printer.Printf("walk_files_per_sec=%.2f\n", res.WalkThroughput)
	ctx.Printer.Printf("stat_p50_us=%.2f\n", res.StatLatency.P50Us)
	ctx.Printer.Printf("stat_p95_us=%.2f\n", res.StatLatency.P95Us)
	ctx.Printer.Printf("stat_p99_us=%.2f\n", res.StatLatency.P99Us)
	ctx.Printer.Printf("total_duration_ms=%.2f\n", float64(res.TotalDuration.Microseconds())/1000.0)
}

// RenderHuman renders an impressive graphical dashboard card with bar charts and percentiles.
func RenderHuman(ctx *command.Context, res Results) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	borderStyle := th.Style(theme.RoleAccent)

	// Throughput visual bar comparison
	maxMBs := math.Max(res.SeqWriteMBs, res.SeqReadMBs)
	if maxMBs <= 0 {
		maxMBs = 1
	}

	writeBarLen := int(math.Round((res.SeqWriteMBs / maxMBs) * 20.0))
	readBarLen := int(math.Round((res.SeqReadMBs / maxMBs) * 20.0))
	if writeBarLen < 1 {
		writeBarLen = 1
	}
	if readBarLen < 1 {
		readBarLen = 1
	}

	writeBar := th.Format(theme.RoleWarning, strings.Repeat("█", writeBarLen), prof)
	readBar := th.Format(theme.RoleSuccess, strings.Repeat("█", readBarLen), prof)

	throughputLines := []string{
		fmt.Sprintf(" %-18s %s %s",
			th.Format(theme.RoleMuted, "Seq Write (1MB chunks):", prof),
			writeBar,
			th.Format(theme.RoleAccent, fmt.Sprintf("%8.2f MB/s", res.SeqWriteMBs), prof),
		),
		fmt.Sprintf(" %-18s %s %s",
			th.Format(theme.RoleMuted, "Seq Read (1MB chunks):", prof),
			readBar,
			th.Format(theme.RoleAccent, fmt.Sprintf("%8.2f MB/s", res.SeqReadMBs), prof),
		),
		strings.Repeat("─", w-4),
		fmt.Sprintf(" %-18s %s",
			th.Format(theme.RoleMuted, "Directory Walk Rate:", prof),
			th.Format(theme.RoleSuccess, fmt.Sprintf("%.0f files/sec", res.WalkThroughput), prof),
		),
		fmt.Sprintf(" %-18s %s",
			th.Format(theme.RoleMuted, "Metadata Stat Rate:", prof),
			th.Format(theme.RoleSuccess, fmt.Sprintf("%.0f ops/sec", res.StatOpsPerSec), prof),
		),
	}

	card1 := renderer.RenderCard("🚀 Filesystem Throughput", throughputLines, w, borderStyle, unicode, prof)
	for _, l := range card1 {
		ctx.Printer.Println(l)
	}
	ctx.Printer.Println()

	// Latency percentiles card
	latValues := []float64{
		res.StatLatency.MinUs,
		res.StatLatency.P50Us,
		res.StatLatency.P90Us,
		res.StatLatency.P95Us,
		res.StatLatency.P99Us,
		res.StatLatency.MaxUs,
	}
	spark := renderer.RenderSparkline(latValues)

	latLines := []string{
		fmt.Sprintf(" %-16s %s (min → p50 → p90 → p95 → p99 → max)", th.Format(theme.RoleMuted, "Curve:", prof), th.Format(theme.RoleAccent, spark, prof)),
		fmt.Sprintf("   • p50 (median): %s     • p95: %s",
			th.Format(theme.RoleSuccess, fmt.Sprintf("%.2f µs", res.StatLatency.P50Us), prof),
			th.Format(theme.RoleWarning, fmt.Sprintf("%.2f µs", res.StatLatency.P95Us), prof),
		),
		fmt.Sprintf("   • p99:          %s     • Max: %s",
			th.Format(theme.RoleError, fmt.Sprintf("%.2f µs", res.StatLatency.P99Us), prof),
			th.Format(theme.RoleError, fmt.Sprintf("%.2f µs", res.StatLatency.MaxUs), prof),
		),
		strings.Repeat("─", w-4),
		fmt.Sprintf(" %-16s %s (%dMB payload, %d stat iterations in %.2fms)",
			th.Format(theme.RoleMuted, "Benchmark Run:", prof),
			th.Format(theme.RoleSuccess, "PASSED", prof),
			res.SizeMB,
			res.Iterations,
			float64(res.TotalDuration.Microseconds())/1000.0,
		),
	}

	card2 := renderer.RenderCard("⏱️  Stat Latency Profile", latLines, w, borderStyle, unicode, prof)
	for _, l := range card2 {
		ctx.Printer.Println(l)
	}
}
