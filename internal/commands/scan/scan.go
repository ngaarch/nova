package scancmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/commands/cat"
	"nova/internal/output"
)

// Finding represents a detected credential or high-entropy secret.
type Finding struct {
	File     string  `json:"file"`
	Line     int     `json:"line"`
	RuleID   string  `json:"rule_id"`
	RuleName string  `json:"rule_name"`
	Severity string  `json:"severity"`
	Masked   string  `json:"masked"`
	Snippet  string  `json:"snippet"`
	Entropy  float64 `json:"entropy,omitempty"`
}

// ScanResult aggregates findings and performance metrics.
type ScanResult struct {
	Target        string        `json:"target"`
	FilesScanned  int           `json:"files_scanned"`
	TotalFindings int           `json:"total_findings"`
	CriticalCount int           `json:"critical_count"`
	HighCount     int           `json:"high_count"`
	MediumCount   int           `json:"medium_count"`
	Duration      time.Duration `json:"duration"`
	Findings      []Finding     `json:"findings"`
}

// Command returns the registered Command instance for scan.
func Command() *command.Command {
	return &command.Command{
		Name:        "scan",
		Aliases:     []string{"audit", "detect", "secrets"},
		Summary:     "Developer secret scanner and Shannon entropy detector",
		Usage:       "nova scan [path] [flags]",
		Description: "Scan directories or files for leaked API keys, tokens, credentials, and high-entropy secrets.",
		Phase:       30,
		Run:         Run,
	}
}

// Run executes the scan command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	start := time.Now()

	res, err := ExecuteScan(opts)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	res.Duration = time.Since(start).Round(time.Millisecond)

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

// ExecuteScan walks target path and evaluates files against rules and entropy checks.
func ExecuteScan(opts Options) (*ScanResult, error) {
	fi, err := os.Stat(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("stat target %q: %w", opts.Target, err)
	}

	res := &ScanResult{
		Target: opts.Target,
	}

	// Filter rules if specified
	var activeRules []Rule
	if opts.Rule != "" {
		targetRule := strings.ToLower(opts.Rule)
		for _, r := range DefaultRules {
			if strings.ToLower(r.ID) == targetRule || strings.Contains(strings.ToLower(r.Name), targetRule) {
				activeRules = append(activeRules, r)
			}
		}
		if len(activeRules) == 0 {
			return nil, fmt.Errorf("unknown secret rule %q (available: github-pat, aws-key, slack-webhook, openai-api-key, stripe-key, private-key, generic-secret)", opts.Rule)
		}
	} else {
		activeRules = DefaultRules
	}

	if !fi.IsDir() {
		// Single file scan
		findings, err := scanFile(opts.Target, activeRules, opts.MinEntropy)
		if err == nil {
			res.FilesScanned = 1
			res.Findings = append(res.Findings, findings...)
		}
	} else {
		// Recursive directory walk
		_ = filepath.Walk(opts.Target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Check ignored paths
			base := info.Name()
			for _, ig := range opts.IgnorePaths {
				if base == ig || strings.Contains(path, string(filepath.Separator)+ig+string(filepath.Separator)) {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if info.IsDir() {
				return nil
			}

			// Skip excessively large files (> 5MB)
			if info.Size() > 5*1024*1024 {
				return nil
			}

			findings, scanErr := scanFile(path, activeRules, opts.MinEntropy)
			if scanErr == nil {
				res.FilesScanned++
				res.Findings = append(res.Findings, findings...)
			}
			return nil
		})
	}

	res.TotalFindings = len(res.Findings)
	for _, f := range res.Findings {
		switch f.Severity {
		case "CRITICAL":
			res.CriticalCount++
		case "HIGH":
			res.HighCount++
		case "MEDIUM":
			res.MediumCount++
		}
	}

	return res, nil
}

func scanFile(path string, rules []Rule, minEntropy float64) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Sample 1KB to check binary
	sample := make([]byte, 1024)
	n, _ := f.Read(sample)
	if cat.IsBinary(sample[:n]) {
		return nil, nil // Skip binary files
	}
	_, _ = f.Seek(0, io.SeekStart)

	var findings []Finding
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		// 1. Rule Pattern matching
		for _, r := range rules {
			loc := r.Regex.FindStringIndex(line)
			if loc != nil {
				match := line[loc[0]:loc[1]]
				// For generic-secret, extract actual value if captured
				if r.ID == "generic-secret" {
					sub := r.Regex.FindStringSubmatch(line)
					if len(sub) >= 3 {
						match = sub[2]
					}
				}

				snippet := makeSnippet(line, loc[0], loc[1])
				findings = append(findings, Finding{
					File:     path,
					Line:     lineNum,
					RuleID:   r.ID,
					RuleName: r.Name,
					Severity: r.Severity,
					Masked:   MaskSecret(match),
					Snippet:  snippet,
				})
				break
			}
		}

		// 2. High Shannon entropy token detection on words > 24 chars
		words := strings.Fields(line)
		for _, w := range words {
			cleanWord := strings.Trim(w, `"'=:,;()[]{}<>`)
			if len(cleanWord) >= 24 && !strings.Contains(cleanWord, "/") && !strings.Contains(cleanWord, "\\") {
				ent := CalculateShannonEntropy(cleanWord)
				if ent >= minEntropy {
					// Check if already matched
					already := false
					for _, fnd := range findings {
						if fnd.Line == lineNum && fnd.File == path {
							already = true
							break
						}
					}
					if !already {
						idx := strings.Index(line, cleanWord)
						snippet := makeSnippet(line, idx, idx+len(cleanWord))
						findings = append(findings, Finding{
							File:     path,
							Line:     lineNum,
							RuleID:   "high-entropy",
							RuleName: "High Shannon Entropy String",
							Severity: "HIGH",
							Masked:   MaskSecret(cleanWord),
							Snippet:  snippet,
							Entropy:  ent,
						})
					}
				}
			}
		}
	}

	return findings, nil
}

func makeSnippet(line string, start, end int) string {
	var buf bytes.Buffer
	clean := strings.TrimSpace(line)
	if len(clean) > 80 {
		return clean[:77] + "..."
	}
	buf.WriteString(clean)
	return buf.String()
}
