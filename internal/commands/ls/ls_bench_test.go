package ls

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"nova/internal/output"
)

func createBenchmarkDir(b *testing.B, count int) string {
	b.Helper()
	dir := b.TempDir()
	for i := 0; i < count; i++ {
		fname := filepath.Join(dir, fmt.Sprintf("file_%06d.txt", i))
		if err := os.WriteFile(fname, []byte("data"), 0o644); err != nil {
			b.Fatalf("failed to create benchmark fixture %d: %v", i, err)
		}
	}
	return dir
}

func BenchmarkLSEmptyDir(b *testing.B) {
	dir := b.TempDir()
	var stdout, stderr bytes.Buffer
	ctx := setupTestContext(&stdout, &stderr, output.ModePlain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = Run(ctx, []string{dir})
	}
}

func BenchmarkLSSmallDirCompact(b *testing.B) {
	dir := createBenchmarkDir(b, 100)
	var stdout, stderr bytes.Buffer
	ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = Run(ctx, []string{dir})
	}
}

func BenchmarkLSSmallDirLong(b *testing.B) {
	dir := createBenchmarkDir(b, 100)
	var stdout, stderr bytes.Buffer
	ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = Run(ctx, []string{"-l", dir})
	}
}

func BenchmarkLSMediumDirCompact(b *testing.B) {
	dir := createBenchmarkDir(b, 1000)
	var stdout, stderr bytes.Buffer
	ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = Run(ctx, []string{dir})
	}
}

func BenchmarkLSMediumDirLong(b *testing.B) {
	dir := createBenchmarkDir(b, 1000)
	var stdout, stderr bytes.Buffer
	ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = Run(ctx, []string{"-l", dir})
	}
}
