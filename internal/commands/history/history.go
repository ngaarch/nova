package historycmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// CommandStats stores analytics for an individual command.
type CommandStats struct {
	Command    string  `json:"command"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Category   string  `json:"category"`
}

// CategorySummary stores aggregated usage for a command category.
type CategorySummary struct {
	Category   string  `json:"category"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// Report holds full shell history analytics.
type Report struct {
	HistoryFile    string            `json:"history_file"`
	TotalCommands  int               `json:"total_commands"`
	UniqueCommands int               `json:"unique_commands"`
	TopCommands    []CommandStats    `json:"top_commands"`
	Categories     []CategorySummary `json:"categories"`
}

// Command returns the registered Command instance for history.
func Command() *command.Command {
	return &command.Command{
		Name:        "history",
		Aliases:     []string{"hist", "analytics"},
		Summary:     "Shell history intelligence and command productivity analytics",
		Usage:       "nova history [filter] [flags]",
		Description: "Analyze past shell executions, identify most-used commands, and breakdown usage by category.",
		Phase:       25,
		Run:         Run,
	}
}

// Run executes the history command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	filePath := opts.File
	if filePath == "" {
		filePath = LocateHistoryFile()
	}

	if filePath == "" {
		return fmt.Errorf("no shell history file found (specify with --file=<path>)")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read history file %s: %w", filePath, err)
	}

	commands := ParseHistory(data)
	report := AnalyzeHistory(commands, opts.Count, opts.Filter, filePath)

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(report)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderPlain(ctx.Stdout, report)
	} else {
		RenderDashboard(ctx.Stdout, report, ctx)
	}
	return nil
}

// LocateHistoryFile searches for default history files in the user profile.
func LocateHistoryFile() string {
	if envHist := os.Getenv("HISTFILE"); envHist != "" {
		if _, err := os.Stat(envHist); err == nil {
			return envHist
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".bash_history"),
		filepath.Join(home, ".local", "share", "fish", "fish_history"),
		filepath.Join(home, ".history"),
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}

	return ""
}

// ParseHistory parses lines from bash, zsh, or fish history.
func ParseHistory(data []byte) []string {
	var commands []string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Zsh extended history: ": 1680000000:0;command"
		if strings.HasPrefix(line, ":") {
			if semi := strings.Index(line, ";"); semi != -1 && semi+1 < len(line) {
				line = strings.TrimSpace(line[semi+1:])
			}
		}

		// Fish history metadata lines
		if strings.HasPrefix(line, "when:") || strings.HasPrefix(line, "paths:") {
			continue
		}

		// Fish history: "- cmd: command"
		if strings.HasPrefix(line, "- cmd: ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "- cmd: "))
		}

		if line != "" {
			commands = append(commands, line)
		}
	}

	return commands
}

// AnalyzeHistory computes statistics, rankings, and category breakdowns.
func AnalyzeHistory(commands []string, topN int, filter string, filePath string) *Report {
	if topN <= 0 {
		topN = 10
	}

	counts := make(map[string]int)
	catCounts := make(map[string]int)
	totalFiltered := 0

	filterLower := strings.ToLower(filter)

	for _, fullCmd := range commands {
		fields := strings.Fields(fullCmd)
		if len(fields) == 0 {
			continue
		}
		rootCmd := fields[0]

		// If path specified, take base name (e.g. ./scripts/run -> run)
		rootCmd = filepath.Base(rootCmd)

		if filterLower != "" && !strings.Contains(strings.ToLower(fullCmd), filterLower) {
			continue
		}

		counts[rootCmd]++
		cat := Categorize(rootCmd)
		catCounts[cat]++
		totalFiltered++
	}

	var stats []CommandStats
	for cmd, cnt := range counts {
		pct := 0.0
		if totalFiltered > 0 {
			pct = (float64(cnt) / float64(totalFiltered)) * 100.0
		}
		stats = append(stats, CommandStats{
			Command:    cmd,
			Count:      cnt,
			Percentage: pct,
			Category:   Categorize(cmd),
		})
	}

	// Sort descending by count
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Command < stats[j].Command
	})

	if len(stats) > topN {
		stats = stats[:topN]
	}

	var categories []CategorySummary
	for cat, cnt := range catCounts {
		pct := 0.0
		if totalFiltered > 0 {
			pct = (float64(cnt) / float64(totalFiltered)) * 100.0
		}
		categories = append(categories, CategorySummary{
			Category:   cat,
			Count:      cnt,
			Percentage: pct,
		})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Count > categories[j].Count
	})

	return &Report{
		HistoryFile:    filePath,
		TotalCommands:  totalFiltered,
		UniqueCommands: len(counts),
		TopCommands:    stats,
		Categories:     categories,
	}
}

// Categorize maps command root to common development categories.
func Categorize(cmd string) string {
	switch strings.ToLower(cmd) {
	case "git", "gh", "glab":
		return "Git"
	case "go", "cargo", "rustc", "npm", "pnpm", "yarn", "bun", "python", "python3", "pip", "node", "make", "gcc", "clang", "java", "mvn", "gradle", "deno", "composer", "php", "ruby":
		return "Development"
	case "cd", "ls", "cat", "rm", "cp", "mv", "mkdir", "chmod", "chown", "kill", "ps", "top", "pwd", "grep", "find", "stat", "du", "df", "which", "touch", "nova":
		return "System"
	case "docker", "docker-compose", "podman", "kubectl", "k9s", "helm", "minikube":
		return "Containers"
	case "curl", "wget", "ssh", "scp", "ping", "netstat", "nmap", "nc", "traceroute", "dig", "nslookup":
		return "Network"
	case "vim", "nvim", "nano", "code", "emacs", "hx":
		return "Editors"
	default:
		return "General"
	}
}
