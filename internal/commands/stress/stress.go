package stresscmd

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/renderer"
)

// WorkloadResult holds metrics for an individual stress workload.
type WorkloadResult struct {
	Name       string  `json:"name"`
	OpsPerSec  float64 `json:"ops_per_sec"`
	Throughput string  `json:"throughput"`
	TotalOps   int64   `json:"total_ops"`
	TotalBytes int64   `json:"total_bytes,omitempty"`
}

// StressResult aggregates overall benchmark metrics across all workloads.
type StressResult struct {
	Threads   int              `json:"threads"`
	Duration  time.Duration    `json:"duration"`
	Workloads []WorkloadResult `json:"workloads"`
}

// Command returns the registered Command instance for stress.
func Command() *command.Command {
	return &command.Command{
		Name:        "stress",
		Aliases:     []string{"cpu", "burn", "benchmark"},
		Summary:     "Multi-core CPU, memory, and cryptographic hashing stress suite",
		Usage:       "nova stress [flags]",
		Description: "Stress-test system hardware and benchmark SHA-256, floating-point math, and memory bandwidth.",
		Phase:       31,
		Run:         Run,
	}
}

// Run executes the stress command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	res := StressResult{
		Threads:  opts.Threads,
		Duration: opts.Duration,
	}

	runAll := opts.Algo == "all" || opts.Algo == ""

	if runAll || opts.Algo == "sha256" {
		res.Workloads = append(res.Workloads, runSHA256Stress(opts.Threads, opts.Duration))
	}
	if runAll || opts.Algo == "math" {
		res.Workloads = append(res.Workloads, runMathStress(opts.Threads, opts.Duration))
	}
	if runAll || opts.Algo == "mem" {
		res.Workloads = append(res.Workloads, runMemStress(opts.Threads, opts.Duration))
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderPlain(ctx.Stdout, res)
	} else {
		RenderDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

func runSHA256Stress(threads int, dur time.Duration) WorkloadResult {
	var totalHashes int64
	var stopFlag int32
	var wg sync.WaitGroup

	chunk := make([]byte, 512)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	start := time.Now()

	for t := 0; t < threads; t++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			localChunk := append([]byte(nil), chunk...)
			localChunk[0] = byte(workerID)
			var count int64
			for atomic.LoadInt32(&stopFlag) == 0 {
				h := sha256.Sum256(localChunk)
				localChunk[1] = h[0]
				count++
				if count%1000 == 0 {
					atomic.AddInt64(&totalHashes, 1000)
					count = 0
				}
			}
			atomic.AddInt64(&totalHashes, count)
		}(t)
	}

	time.Sleep(dur)
	atomic.StoreInt32(&stopFlag, 1)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}

	hashes := atomic.LoadInt64(&totalHashes)
	opsPerSec := float64(hashes) / elapsed
	bytesPerSec := int64(opsPerSec * 512)

	return WorkloadResult{
		Name:       "SHA-256",
		TotalOps:   hashes,
		OpsPerSec:  opsPerSec,
		TotalBytes: hashes * 512,
		Throughput: fmt.Sprintf("%s/s", renderer.FormatSize(bytesPerSec, true)),
	}
}

func runMathStress(threads int, dur time.Duration) WorkloadResult {
	var totalOps int64
	var stopFlag int32
	var wg sync.WaitGroup

	start := time.Now()

	for t := 0; t < threads; t++ {
		wg.Add(1)
		go func(seed float64) {
			defer wg.Done()
			x := seed + 1.5
			var count int64
			for atomic.LoadInt32(&stopFlag) == 0 {
				// Trigonometric and algebraic operations
				x = math.Sqrt(x*x + 1.0)
				x = math.Sin(x) + math.Cos(x)
				if x > 1000.0 {
					x = 1.5
				}
				count++
				if count%5000 == 0 {
					atomic.AddInt64(&totalOps, 5000*4) // 4 ops per cycle
					count = 0
				}
			}
			atomic.AddInt64(&totalOps, count*4)
		}(float64(t))
	}

	time.Sleep(dur)
	atomic.StoreInt32(&stopFlag, 1)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}

	ops := atomic.LoadInt64(&totalOps)
	opsPerSec := float64(ops) / elapsed
	mflops := opsPerSec / 1_000_000.0

	return WorkloadResult{
		Name:       "Floating Math",
		TotalOps:   ops,
		OpsPerSec:  opsPerSec,
		Throughput: fmt.Sprintf("%.2f MFLOPS", mflops),
	}
}

func runMemStress(threads int, dur time.Duration) WorkloadResult {
	var totalAllocs int64
	var stopFlag int32
	var wg sync.WaitGroup

	const blockSize = 64 * 1024 // 64 KB
	start := time.Now()

	for t := 0; t < threads; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bufA := make([]byte, blockSize)
			bufB := make([]byte, blockSize)
			for i := range bufA {
				bufA[i] = byte(i)
			}
			var count int64
			for atomic.LoadInt32(&stopFlag) == 0 {
				copy(bufB, bufA)
				bufB[0]++
				count++
				if count%1000 == 0 {
					atomic.AddInt64(&totalAllocs, 1000)
					count = 0
				}
			}
			atomic.AddInt64(&totalAllocs, count)
		}()
	}

	time.Sleep(dur)
	atomic.StoreInt32(&stopFlag, 1)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}

	allocs := atomic.LoadInt64(&totalAllocs)
	opsPerSec := float64(allocs) / elapsed
	bytesPerSec := int64(opsPerSec * blockSize)

	return WorkloadResult{
		Name:       "Memory Bandwidth",
		TotalOps:   allocs,
		OpsPerSec:  opsPerSec,
		TotalBytes: allocs * blockSize,
		Throughput: fmt.Sprintf("%s/s", renderer.FormatSize(bytesPerSec, true)),
	}
}
