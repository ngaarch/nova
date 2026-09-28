package hash

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// Result represents the hashing outcome for a single file.
type Result struct {
	File      string `json:"file"`
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	Error     string `json:"error,omitempty"`
}

// CheckResult represents verification of a file against expected checksum.
type CheckResult struct {
	File      string `json:"file"`
	Algorithm string `json:"algorithm"`
	Expected  string `json:"expected"`
	Actual    string `json:"actual,omitempty"`
	Status    string `json:"status"` // PASS, FAILED, MISSING
	Error     string `json:"error,omitempty"`
}

// CheckSummary aggregates results of verification.
type CheckSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Missing int `json:"missing"`
}

// Command returns the registered Command instance for hash.
func Command() *command.Command {
	return &command.Command{
		Name:        "hash",
		Aliases:     []string{"checksum", "digest", "sha256"},
		Summary:     "Compute and verify cryptographic and cyclic checksums (SHA256, SHA512, MD5, CRC32)",
		Usage:       "nova hash [flags] [files...]",
		Description: "Calculate cryptographic digests, verify integrity files, and inspect checksums with standard format compatibility.",
		Phase:       17,
		Run:         Run,
	}
}

// Run executes the hash command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	// Validate algorithm
	algo := strings.ToLower(opts.Algorithm)
	switch algo {
	case "sha256", "sha512", "sha1", "md5", "crc32":
		// valid
	default:
		return fmt.Errorf("unsupported hash algorithm %q (supported: sha256, sha512, sha1, md5, crc32)", algo)
	}

	// Mode 1: Check / verify checksum file
	if opts.CheckFile != "" {
		return runCheck(ctx, opts)
	}

	// Mode 2: Compute hashes
	return runCompute(ctx, opts)
}

func newHasher(algo string) (hash.Hash, error) {
	switch strings.ToLower(algo) {
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	case "crc32":
		return hash.Hash(crc32.NewIEEE()), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}

// HashReader computes the hash of reader data.
func HashReader(r io.Reader, algo string) (string, int64, error) {
	h, err := newHasher(algo)
	if err != nil {
		return "", 0, err
	}

	n, err := io.Copy(h, r)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// HashFile computes the hash of a file on disk.
func HashFile(path string, algo string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	return HashReader(f, algo)
}

func runCompute(ctx *command.Context, opts Options) error {
	targets := opts.Paths
	if len(targets) == 0 {
		targets = []string{"-"}
	}

	var fileList []string
	for _, target := range targets {
		if target == "-" {
			fileList = append(fileList, target)
			continue
		}

		info, err := os.Stat(target)
		if err != nil {
			// Record file even if stat fails so error is reported
			fileList = append(fileList, target)
			continue
		}

		if info.IsDir() {
			if !opts.Recursive {
				return fmt.Errorf("%s: is a directory (use -r / --recursive to hash directories)", target)
			}
			_ = filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				fileList = append(fileList, path)
				return nil
			})
		} else {
			fileList = append(fileList, target)
		}
	}

	var results []Result
	for _, target := range fileList {
		var digest string
		var size int64
		var err error

		if target == "-" {
			digest, size, err = HashReader(ctx.Stdin, opts.Algorithm)
			target = "<stdin>"
		} else {
			digest, size, err = HashFile(target, opts.Algorithm)
		}

		res := Result{
			File:      target,
			Algorithm: strings.ToUpper(opts.Algorithm),
			Size:      size,
		}
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Hash = digest
		}
		results = append(results, res)
	}

	return RenderCompute(ctx, results, opts)
}

func runCheck(ctx *command.Context, opts Options) error {
	var reader io.Reader
	if opts.CheckFile == "-" {
		reader = ctx.Stdin
	} else {
		f, err := os.Open(opts.CheckFile)
		if err != nil {
			return fmt.Errorf("failed to open check file %s: %w", opts.CheckFile, err)
		}
		defer f.Close()
		reader = f
	}

	scanner := bufio.NewScanner(reader)
	var checks []CheckResult
	summary := CheckSummary{}

	checkDir := "."
	if opts.CheckFile != "-" {
		checkDir = filepath.Dir(opts.CheckFile)
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Standard format: <hash>  <filename> or <hash> *<filename>
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		expectedHash := strings.ToLower(fields[0])
		targetFile := strings.TrimPrefix(fields[1], "*")

		// If path is relative, resolve relative to check file dir
		resolvedPath := targetFile
		if !filepath.IsAbs(resolvedPath) && checkDir != "." {
			resolvedPath = filepath.Join(checkDir, resolvedPath)
		}

		// Detect algorithm based on expected hash length if default
		algo := opts.Algorithm
		switch len(expectedHash) {
		case 8:
			algo = "crc32"
		case 32:
			algo = "md5"
		case 40:
			algo = "sha1"
		case 64:
			algo = "sha256"
		case 128:
			algo = "sha512"
		}

		checkRes := CheckResult{
			File:      targetFile,
			Algorithm: strings.ToUpper(algo),
			Expected:  expectedHash,
		}

		actualHash, _, err := HashFile(resolvedPath, algo)
		if err != nil {
			checkRes.Status = "MISSING"
			checkRes.Error = err.Error()
			summary.Missing++
		} else {
			checkRes.Actual = actualHash
			if strings.EqualFold(actualHash, expectedHash) {
				checkRes.Status = "PASS"
				summary.Passed++
			} else {
				checkRes.Status = "FAILED"
				summary.Failed++
			}
		}

		summary.Total++
		checks = append(checks, checkRes)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading check file: %w", err)
	}

	return RenderCheck(ctx, checks, summary, opts)
}
