package cat

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/output"
	"nova/internal/terminal"
)

func BenchmarkCatPlainStreaming(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_10k.txt")
	var buf strings.Builder
	for i := 0; i < 10000; i++ {
		buf.WriteString("Benchmark streaming line for memory bound evaluation\n")
	}
	os.WriteFile(path, []byte(buf.String()), 0644)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
		ctx.Stdout = io.Discard
		if err := Run(ctx, []string{path}); err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkCatSyntaxHighlighting(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.go")
	sampleCode := `package main

import (
	"fmt"
	"time"
)

func ProcessData(items []string) error {
	for i, item := range items {
		// Log progress
		fmt.Printf("[%d] Item: %s\n", i, item)
	}
	return nil
}
`
	var buf strings.Builder
	for i := 0; i < 500; i++ {
		buf.WriteString(sampleCode)
	}
	os.WriteFile(path, []byte(buf.String()), 0644)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, _, _ := newTestContext("", output.ModeHuman, terminal.ColorTrueColor)
		ctx.Stdout = io.Discard
		ctx.Caps.Height = 0 // Disable pager for pure highlighting throughput
		if err := Run(ctx, []string{"--no-pager", path}); err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkFormatHexDump(b *testing.B) {
	data := bytes.Repeat([]byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00}, 128) // 1024 bytes
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = FormatHexDump(data, 0)
	}
}
