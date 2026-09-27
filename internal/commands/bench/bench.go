package bench

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// Metric holds an operation measurement.
type Metric struct {
	Name       string        `json:"name"`
	Throughput float64       `json:"throughput_mb_s,omitempty"` // MB/s
	Rate       float64       `json:"ops_per_sec,omitempty"`     // ops/s
	Duration   time.Duration `json:"duration"`
}

// LatencyPercentiles holds latency percentiles in microseconds.
type LatencyPercentiles struct {
	MinUs float64 `json:"min_us"`
	P50Us float64 `json:"p50_us"`
	P90Us float64 `json:"p90_us"`
	P95Us float64 `json:"p95_us"`
	P99Us float64 `json:"p99_us"`
	MaxUs float64 `json:"max_us"`
}

// Results aggregates all benchmark outcomes.
type Results struct {
	ScratchDir    string             `json:"scratch_dir"`
	SizeMB        int                `json:"size_mb"`
	Iterations    int                `json:"iterations"`
	SeqWriteMBs   float64            `json:"seq_write_mb_s"`
	SeqReadMBs    float64            `json:"seq_read_mb_s"`
	StatOpsPerSec float64            `json:"stat_ops_per_sec"`
	WalkThroughput float64           `json:"walk_files_per_sec"`
	StatLatency   LatencyPercentiles `json:"stat_latency_us"`
	TotalDuration time.Duration      `json:"total_duration"`
}

// Command returns the registered Command instance for the bench subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "bench",
		Aliases:     []string{"benchmark"},
		Summary:     "Benchmark filesystem I/O throughput, directory walking, and stat latency",
		Usage:       "nova bench [flags]",
		Description: "Measure disk sequential read/write speed, file metadata lookup latency, and tree traversal rates.",
		Phase:       14,
		Run:         Run,
	}
}

// Run executes the bench command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.Plain {
		ctx.Printer.Mode = output.ModePlain
	} else if opts.JSON {
		ctx.Printer.Mode = output.ModeJSON
	}

	results, err := runBenchmarkSuite(ctx, opts)
	if err != nil {
		return err
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, results)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, results)
	} else {
		RenderHuman(ctx, results)
	}

	return nil
}

func runBenchmarkSuite(ctx *command.Context, opts Options) (Results, error) {
	scratchDir, err := os.MkdirTemp("", "nova_bench_*")
	if err != nil {
		return Results{}, fmt.Errorf("failed to create temp benchmark dir: %w", err)
	}
	defer os.RemoveAll(scratchDir)

	totalStart := time.Now()

	// 1. Sequential Write Benchmark
	chunkSize := 1024 * 1024 // 1MB buffer
	buffer := make([]byte, chunkSize)
	rand.Read(buffer[:1024]) // seed beginning

	targetFile := filepath.Join(scratchDir, "bench_payload.bin")
	f, err := os.Create(targetFile)
	if err != nil {
		return Results{}, fmt.Errorf("failed to create bench target file: %w", err)
	}

	writeStart := time.Now()
	for i := 0; i < opts.SizeMB; i++ {
		if _, wErr := f.Write(buffer); wErr != nil {
			f.Close()
			return Results{}, fmt.Errorf("write error: %w", wErr)
		}
	}
	if sErr := f.Sync(); sErr != nil {
		f.Close()
		return Results{}, fmt.Errorf("sync error: %w", sErr)
	}
	writeDuration := time.Since(writeStart)
	f.Close()

	seqWriteMBs := float64(opts.SizeMB) / writeDuration.Seconds()

	// 2. Sequential Read Benchmark
	rf, err := os.Open(targetFile)
	if err != nil {
		return Results{}, fmt.Errorf("failed to open bench file for read: %w", err)
	}

	readBuffer := make([]byte, chunkSize)
	readStart := time.Now()
	var totalRead int64
	for {
		n, rErr := rf.Read(readBuffer)
		totalRead += int64(n)
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			rf.Close()
			return Results{}, fmt.Errorf("read error: %w", rErr)
		}
	}
	readDuration := time.Since(readStart)
	rf.Close()

	seqReadMBs := (float64(totalRead) / (1024 * 1024)) / readDuration.Seconds()

	// 3. Stat Latency Benchmark
	var latencies []float64
	statStart := time.Now()
	for i := 0; i < opts.Iterations; i++ {
		t0 := time.Now()
		_, _ = os.Stat(targetFile)
		latUs := float64(time.Since(t0).Nanoseconds()) / 1000.0
		latencies = append(latencies, latUs)
	}
	statDuration := time.Since(statStart)
	statOpsPerSec := float64(opts.Iterations) / statDuration.Seconds()

	sort.Float64s(latencies)
	percentiles := LatencyPercentiles{
		MinUs: latencies[0],
		P50Us: latencies[len(latencies)*50/100],
		P90Us: latencies[len(latencies)*90/100],
		P95Us: latencies[len(latencies)*95/100],
		P99Us: latencies[len(latencies)*99/100],
		MaxUs: latencies[len(latencies)-1],
	}

	// 4. Directory Traversal Walk Benchmark
	// Create small tree of files
	walkDir := filepath.Join(scratchDir, "tree")
	os.MkdirAll(walkDir, 0755)
	fileCount := 200
	for i := 0; i < fileCount; i++ {
		subDir := filepath.Join(walkDir, fmt.Sprintf("sub_%d", i%10))
		os.MkdirAll(subDir, 0755)
		dummyFile := filepath.Join(subDir, fmt.Sprintf("file_%d.txt", i))
		os.WriteFile(dummyFile, []byte("ok"), 0644)
	}

	walkStart := time.Now()
	var walkedCount int
	_ = filepath.Walk(walkDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() {
			walkedCount++
		}
		return nil
	})
	walkDuration := time.Since(walkStart)
	walkThroughput := float64(walkedCount) / walkDuration.Seconds()

	totalDuration := time.Since(totalStart)

	return Results{
		ScratchDir:     scratchDir,
		SizeMB:         opts.SizeMB,
		Iterations:     opts.Iterations,
		SeqWriteMBs:    seqWriteMBs,
		SeqReadMBs:     seqReadMBs,
		StatOpsPerSec:  statOpsPerSec,
		WalkThroughput: walkThroughput,
		StatLatency:    percentiles,
		TotalDuration:  totalDuration,
	}, nil
}
