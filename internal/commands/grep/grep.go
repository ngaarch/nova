package grep

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// MatchLine represents a matched line within a file.
type MatchLine struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	IsMatch    bool   `json:"is_match"` // true for match, false for context
	MatchStart int    `json:"match_start,omitempty"`
	MatchEnd   int    `json:"match_end,omitempty"`
}

// FileResult holds all matches found within a single file.
type FileResult struct {
	Path       string      `json:"path"`
	Matches    []MatchLine `json:"matches"`
	MatchCount int         `json:"match_count"`
}

// Summary holds aggregate metrics for the grep search.
type Summary struct {
	TotalMatches int           `json:"total_matches"`
	TotalFiles   int           `json:"total_files"`
	FilesScanned int           `json:"files_scanned"`
	Duration     time.Duration `json:"duration"`
}

// Command returns the registered Command instance for the grep subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "grep",
		Aliases:     []string{"search"},
		Summary:     "Fast concurrent code and text search with syntax and regex matching",
		Usage:       "nova grep [flags] <pattern> [paths...]",
		Description: "Recursively search file contents for regular expressions or literal strings.",
		Phase:       13,
		Run:         Run,
	}
}

// Run executes the grep command.
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

	if opts.Query == "" {
		return command.NewUsageError("missing search pattern", "Usage: nova grep [flags] <pattern> [paths...]")
	}

	pattern := opts.Query
	if opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return command.NewUsageError(fmt.Sprintf("invalid regular expression: %v", err), "Check your search regex syntax.")
	}

	startTime := time.Now()
	results, filesScanned, err := executeSearch(opts, re)
	if err != nil {
		return err
	}
	duration := time.Since(startTime)

	totalMatches := 0
	for _, fr := range results {
		totalMatches += fr.MatchCount
	}

	summary := Summary{
		TotalMatches: totalMatches,
		TotalFiles:   len(results),
		FilesScanned: filesScanned,
		Duration:     duration,
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, results, summary)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, results, opts)
	} else {
		RenderHuman(ctx, results, summary, opts, re)
	}

	if totalMatches == 0 {
		return errors.New("no matches found")
	}

	return nil
}

func executeSearch(opts Options, re *regexp.Regexp) ([]FileResult, int, error) {
	var targetFiles []string
	var scannedCount int

	for _, root := range opts.Paths {
		info, err := os.Stat(root)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if matchesExt(root, opts.Extensions) {
				targetFiles = append(targetFiles, root)
			}
			continue
		}

		err = filepath.Walk(root, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			// Skip hidden files unless requested
			if !opts.Hidden && path != root {
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			// Skip common vendor/build directories
			if fi.IsDir() {
				base := filepath.Base(path)
				if base == ".git" || base == "node_modules" || base == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}

			if fi.Mode().IsRegular() {
				if matchesExt(path, opts.Extensions) {
					targetFiles = append(targetFiles, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
	}

	scannedCount = len(targetFiles)
	if scannedCount == 0 {
		return nil, 0, nil
	}

	// Concurrency worker pool
	numWorkers := 8
	if scannedCount < numWorkers {
		numWorkers = scannedCount
	}

	jobs := make(chan string, scannedCount)
	resultsChan := make(chan *FileResult, scannedCount)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				fr := searchInFile(path, re, opts)
				if fr != nil && fr.MatchCount > 0 {
					resultsChan <- fr
				}
			}
		}()
	}

	for _, f := range targetFiles {
		jobs <- f
	}
	close(jobs)

	wg.Wait()
	close(resultsChan)

	var allResults []FileResult
	for fr := range resultsChan {
		allResults = append(allResults, *fr)
	}

	return allResults, scannedCount, nil
}

func matchesExt(path string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	fileExt := strings.ToLower(filepath.Ext(path))
	for _, ext := range exts {
		if fileExt == ext {
			return true
		}
	}
	return false
}

func isBinary(data []byte) bool {
	checkLen := len(data)
	if checkLen > 512 {
		checkLen = 512
	}
	return bytes.IndexByte(data[:checkLen], 0) != -1
}

func searchInFile(path string, re *regexp.Regexp, opts Options) *FileResult {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	// Read initial header to detect binary files
	header := make([]byte, 512)
	n, _ := f.Read(header)
	if isBinary(header[:n]) {
		return nil
	}

	// Rewind to beginning
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil
	}

	scanner := bufio.NewScanner(f)
	var allLines []string
	var matchIndices []int

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		allLines = append(allLines, line)

		if re.MatchString(line) {
			matchIndices = append(matchIndices, lineNum-1)
			if opts.MaxCount > 0 && len(matchIndices) >= opts.MaxCount {
				break
			}
		}
	}

	if len(matchIndices) == 0 {
		return nil
	}

	fr := &FileResult{
		Path:       path,
		MatchCount: len(matchIndices),
	}

	if opts.FilesWithMatch || opts.CountOnly {
		return fr
	}

	// Collect matches with context
	includedLines := make(map[int]bool)
	for _, idx := range matchIndices {
		start := idx - opts.ContextBefore
		if start < 0 {
			start = 0
		}
		end := idx + opts.ContextAfter
		if end >= len(allLines) {
			end = len(allLines) - 1
		}
		for i := start; i <= end; i++ {
			includedLines[i] = true
		}
	}

	for i := 0; i < len(allLines); i++ {
		if !includedLines[i] {
			continue
		}
		isMatch := re.MatchString(allLines[i])
		mLine := MatchLine{
			LineNumber: i + 1,
			Content:    allLines[i],
			IsMatch:    isMatch,
		}
		if isMatch {
			loc := re.FindStringIndex(allLines[i])
			if len(loc) == 2 {
				mLine.MatchStart = loc[0]
				mLine.MatchEnd = loc[1]
			}
		}
		fr.Matches = append(fr.Matches, mLine)
	}

	return fr
}
