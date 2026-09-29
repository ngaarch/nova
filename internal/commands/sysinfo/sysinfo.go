package sysinfo

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/git"
	"nova/internal/output"
)

// BatteryInfo holds battery telemetry for mobile and laptop devices.
type BatteryInfo struct {
	Available bool   `json:"available"`
	Percent   int    `json:"percent,omitempty"`
	Status    string `json:"status,omitempty"`
}

// Info collects complete system, runtime, and environment telemetry.
type Info struct {
	Hostname       string      `json:"hostname"`
	DeviceType     string      `json:"device_type"`
	IsTermux       bool        `json:"is_termux"`
	Battery        BatteryInfo `json:"battery"`
	OS             string      `json:"os"`
	Arch           string      `json:"arch"`
	CPUs           int         `json:"cpus"`
	GoVersion      string      `json:"go_version"`
	Goroutines     int         `json:"goroutines"`
	MemoryAlloc    uint64      `json:"memory_alloc"`
	MemoryTotal    uint64      `json:"memory_total"`
	MemorySys      uint64      `json:"memory_sys"`
	HeapAlloc      uint64      `json:"heap_alloc"`
	NumGC          uint32      `json:"num_gc"`
	Cwd            string      `json:"cwd"`
	Disk           DiskSpace   `json:"disk"`
	TermWidth      int         `json:"term_width"`
	TermHeight     int         `json:"term_height"`
	TermIsTTY      bool        `json:"term_is_tty"`
	TermColor      string      `json:"term_color"`
	TermUnicode    bool        `json:"term_unicode"`
	ThemeName      string      `json:"theme_name"`
	IconMode       string      `json:"icon_mode"`
	GitRepo        bool        `json:"git_repo"`
	GitBranch      string      `json:"git_branch,omitempty"`
	GitClean       bool        `json:"git_clean,omitempty"`
	PathTotalCount int         `json:"path_total_count"`
	PathValidCount int         `json:"path_valid_count"`
	PathDirs       []string    `json:"path_dirs,omitempty"`
}

// Command returns the registered Command instance for the sysinfo subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "sysinfo",
		Aliases:     []string{"sys", "info", "system"},
		Summary:     "Display system, environment, runtime, and terminal diagnostic dashboard",
		Usage:       "nova sysinfo [flags]",
		Description: "Inspect hardware, operating system, Go runtime, memory, storage, and Nova environment.",
		Phase:       14,
		Run:         Run,
	}
}

// Run executes the sysinfo command.
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

	info := gatherInfo(ctx, opts)

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, info)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, info)
	} else {
		RenderHuman(ctx, info, opts)
	}

	return nil
}

func gatherInfo(ctx *command.Context, opts Options) Info {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	disk := getDiskSpace(cwd)

	// PATH analysis
	pathEnv := os.Getenv("PATH")
	dirs := filepath.SplitList(pathEnv)
	validCount := 0
	for _, d := range dirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			validCount++
		}
	}

	// Git inspection
	var gitRepo bool
	var gitBranch string
	var gitClean bool

	repoStatus, err := git.GetRepoStatus(cwd, 100*time.Millisecond)
	if err == nil && repoStatus != nil {
		gitRepo = true
		gitBranch = repoStatus.Branch
		gitClean = len(repoStatus.Statuses) == 0
	}

	devType, isTermux := detectDeviceType()
	battery := getBatteryInfo()

	info := Info{
		Hostname:       hostname,
		DeviceType:     devType,
		IsTermux:       isTermux,
		Battery:        battery,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		CPUs:           runtime.NumCPU(),
		GoVersion:      runtime.Version(),
		Goroutines:     runtime.NumGoroutine(),
		MemoryAlloc:    m.Alloc,
		MemoryTotal:    m.TotalAlloc,
		MemorySys:      m.Sys,
		HeapAlloc:      m.HeapAlloc,
		NumGC:          m.NumGC,
		Cwd:            cwd,
		Disk:           disk,
		TermWidth:      ctx.Caps.Width,
		TermHeight:     ctx.Caps.Height,
		TermIsTTY:      ctx.Caps.IsTTY,
		TermColor:      ctx.Caps.ColorProfile.String(),
		TermUnicode:    ctx.Caps.UnicodeSupported,
		ThemeName:      ctx.Theme.Name,
		IconMode:       ctx.Config.IconMode,
		GitRepo:        gitRepo,
		GitBranch:      gitBranch,
		GitClean:       gitClean,
		PathTotalCount: len(dirs),
		PathValidCount: validCount,
	}

	if opts.Full {
		info.PathDirs = dirs
	}

	return info
}

func detectDeviceType() (string, bool) {
	isTermux := os.Getenv("TERMUX_VERSION") != "" ||
		strings.Contains(os.Getenv("PREFIX"), "com.termux") ||
		strings.Contains(os.Getenv("HOME"), "com.termux")

	if isTermux {
		return "Android (Termux)", true
	}

	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "macOS (Apple Silicon)", false
		}
		return "macOS (Intel)", false
	case "windows":
		return "Windows PC", false
	case "linux":
		return "Linux Host", false
	case "freebsd":
		return "FreeBSD Host", false
	default:
		return runtime.GOOS + "/" + runtime.GOARCH, false
	}
}

func getBatteryInfo() BatteryInfo {
	paths := []string{
		"/sys/class/power_supply/battery",
		"/sys/class/power_supply/BAT0",
		"/sys/class/power_supply/BAT1",
	}
	for _, base := range paths {
		capFile := filepath.Join(base, "capacity")
		if data, err := os.ReadFile(capFile); err == nil {
			pct, _ := strconv.Atoi(strings.TrimSpace(string(data)))
			statFile := filepath.Join(base, "status")
			status := "Unknown"
			if sData, err := os.ReadFile(statFile); err == nil {
				status = strings.TrimSpace(string(sData))
			}
			return BatteryInfo{
				Available: true,
				Percent:   pct,
				Status:    status,
			}
		}
	}
	return BatteryInfo{Available: false}
}
