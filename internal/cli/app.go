package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"nova/internal/commands/cat"
	"nova/internal/commands/ls"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// App manages command registration, flag parsing, context configuration, and execution routing.
type App struct {
	commands []*Command
}

// NewApp initializes an App instance configured with the roadmap command suite.
func NewApp() *App {
	app := &App{}
	app.registerRoadmapCommands()
	return app
}

// registerRoadmapCommands registers all planned v1 subcommands with their scheduled phases.
func (a *App) registerRoadmapCommands() {
	a.commands = []*Command{
		ls.Command(),
		cat.Command(),
		{
			Name:    "tree",
			Summary: "Display directory hierarchy as a visual tree with metrics",
			Phase:   5,
			Run:     nil, // Scheduled for Phase 5
		},
		{
			Name:    "find",
			Summary: "Search files across directories by predicates",
			Phase:   5,
			Run:     nil, // Scheduled for Phase 5
		},
		{
			Name:    "stat",
			Summary: "Display structured file status and filesystem metadata",
			Phase:   5,
			Run:     nil, // Scheduled for Phase 5
		},
		{
			Name:    "du",
			Summary: "Summarize disk usage by directories with human-readable units",
			Phase:   5,
			Run:     nil, // Scheduled for Phase 5
		},
		{
			Name:    "cp",
			Aliases: []string{"copy"},
			Summary: "Copy files and directories with progress, metadata preservation, and safety",
			Phase:   6,
			Run:     nil, // Scheduled for Phase 6
		},
		{
			Name:    "mv",
			Aliases: []string{"move"},
			Summary: "Move and rename files and directories safely",
			Phase:   6,
			Run:     nil, // Scheduled for Phase 6
		},
		{
			Name:    "rm",
			Aliases: []string{"remove"},
			Summary: "Safely delete files and directories with protection traps and dry-run",
			Phase:   6,
			Run:     nil, // Scheduled for Phase 6
		},
		{
			Name:    "mkdir",
			Summary: "Create directories with hierarchy support (-p)",
			Phase:   6,
			Run:     nil, // Scheduled for Phase 6
		},
		{
			Name:    "interactive",
			Aliases: []string{"ui", "tui"},
			Summary: "Interactive terminal file navigator and previewer",
			Phase:   7,
			Run:     nil, // Scheduled for Phase 7
		},
	}
}

// RegisterCommand adds or replaces a command in the app.
func (a *App) RegisterCommand(cmd *Command) {
	for i, existing := range a.commands {
		if existing.Name == cmd.Name {
			a.commands[i] = cmd
			return
		}
	}
	a.commands = append(a.commands, cmd)
}

// parsedFlags holds global flag evaluations.
type parsedFlags struct {
	help       bool
	version    bool
	plain      bool
	json       bool
	debug      bool
	colorMode  string
	theme      string
	iconMode   string
	remainArgs []string
}

// parseGlobalFlags extracts global flags from args, returning remaining arguments.
func parseGlobalFlags(args []string) parsedFlags {
	var pf parsedFlags
	var remain []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			pf.help = true
			i++
		} else if arg == "--version" || arg == "-v" {
			pf.version = true
			i++
		} else if arg == "--plain" {
			pf.plain = true
			i++
		} else if arg == "--json" {
			pf.json = true
			i++
		} else if arg == "--debug" {
			pf.debug = true
			i++
		} else if strings.HasPrefix(arg, "--color=") {
			pf.colorMode = strings.TrimPrefix(arg, "--color=")
			i++
		} else if arg == "--color" && i+1 < len(args) {
			pf.colorMode = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--theme=") {
			pf.theme = strings.TrimPrefix(arg, "--theme=")
			i++
		} else if arg == "--theme" && i+1 < len(args) {
			pf.theme = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--icons=") {
			pf.iconMode = strings.TrimPrefix(arg, "--icons=")
			i++
		} else if arg == "--icons" && i+1 < len(args) {
			pf.iconMode = args[i+1]
			i += 2
		} else {
			remain = append(remain, arg)
			i++
		}
	}
	pf.remainArgs = remain
	return pf
}

// Run executes the application with provided arguments and I/O streams, returning the process exit code.
func (a *App) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "nova: error loading configuration: %v\n", err)
		return ExitFailure
	}

	flags := parseGlobalFlags(args)

	// Apply flag overrides to configuration
	if flags.theme != "" {
		cfg.Theme = flags.theme
	}
	if flags.colorMode != "" {
		cfg.ColorMode = flags.colorMode
	}
	if flags.iconMode != "" {
		cfg.IconMode = flags.iconMode
	}
	if flags.debug {
		cfg.Debug = true
	}

	logger := logging.New(stderr, cfg.Debug)
	logger.Debugf("nova initialized with args: %v", args)

	// Detect terminal capabilities
	var caps terminal.Capabilities
	if f, ok := stdout.(*os.File); ok {
		caps = terminal.Detect(f)
	} else {
		caps = terminal.DetectWithEnv(nil, os.Getenv)
	}

	// Resolve output mode: flags > config > TTY auto-detection
	effectiveMode := output.ResolveMode(flags.plain, flags.json, caps.IsTTY)
	if !flags.plain && !flags.json {
		if flags.colorMode == "always" || cfg.ColorMode == "always" {
			effectiveMode = output.ModeHuman
		} else if cfg.OutputMode != "auto" {
			switch cfg.OutputMode {
			case "json":
				effectiveMode = output.ModeJSON
			case "plain":
				effectiveMode = output.ModePlain
			case "human":
				effectiveMode = output.ModeHuman
			}
		}
	}

	// Apply color overrides
	if flags.colorMode == "never" || cfg.ColorMode == "never" {
		caps.ColorProfile = terminal.ColorNone
	} else if flags.colorMode == "always" || cfg.ColorMode == "always" {
		if caps.ColorProfile == terminal.ColorNone {
			caps.ColorProfile = terminal.Color256
		}
	}

	th := theme.Get(cfg.Theme)
	ctx := NewContext(stdin, stdout, stderr, cfg, caps, th, effectiveMode, logger)

	// Handle global --version
	if flags.version {
		if ctx.Printer.Mode == output.ModeJSON {
			if err := ctx.Printer.PrintJSON(GetVersionInfo()); err != nil {
				ctx.Printer.Errorf("nova: failed to format version JSON: %v", err)
				return ExitFailure
			}
			return ExitSuccess
		}
		ctx.Printer.Println(GetVersionInfo().Format())
		return ExitSuccess
	}

	// Handle global --help or no arguments
	if flags.help || len(flags.remainArgs) == 0 {
		if err := PrintHelp(ctx, a.commands); err != nil {
			ctx.Printer.Errorf("nova: error rendering help: %v", err)
			return ExitFailure
		}
		return ExitSuccess
	}

	// Route to subcommand
	cmdName := flags.remainArgs[0]
	cmdArgs := flags.remainArgs[1:]

	var targetCmd *Command
	for _, cmd := range a.commands {
		if cmd.Matches(cmdName) {
			targetCmd = cmd
			break
		}
	}

	if targetCmd == nil {
		err := NewUsageError(
			fmt.Sprintf("unknown command %q", cmdName),
			"Run 'nova --help' to inspect available commands.",
		)
		ctx.Printer.Error(err.Error())
		return ExitUsage
	}

	// Unimplemented command handler (scheduled for later roadmap phase)
	if targetCmd.Run == nil {
		err := NewOpError(
			fmt.Sprintf("command %q is not implemented yet (scheduled for Phase %d in ROADMAP.md)", targetCmd.Name, targetCmd.Phase),
			"",
			nil,
			fmt.Sprintf("Check 'ROADMAP.md' or run 'nova --help' for available capabilities in the current release."),
		)
		ctx.Printer.Error(err.Error())
		return ExitFailure
	}

	// Execute command
	if err := targetCmd.Run(ctx, cmdArgs); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			ctx.Printer.Error(exitErr.Error())
			return exitErr.Code
		}
		ctx.Printer.Errorf("nova: error executing %s: %v", targetCmd.Name, err)
		return ExitFailure
	}

	return ExitSuccess
}
