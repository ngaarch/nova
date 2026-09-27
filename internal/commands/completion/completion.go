package completion

import (
	"fmt"
	"strings"

	"nova/internal/command"
)

// Command returns the registered Command instance for completion.
func Command() *command.Command {
	return &command.Command{
		Name:        "completion",
		Aliases:     []string{},
		Summary:     "Generate shell autocompletion script (bash, zsh, fish)",
		Usage:       "nova completion [bash|zsh|fish]",
		Description: "Outputs shell completion code for the specified shell to integrate into shell profiles.",
		Phase:       11,
		Run:         Run,
	}
}

// Run executes the completion command.
func Run(ctx *command.Context, args []string) error {
	shell := "bash"
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") && arg != "" {
			shell = strings.ToLower(arg)
			break
		}
	}

	switch shell {
	case "bash":
		ctx.Printer.Println(BashCompletion())
	case "zsh":
		ctx.Printer.Println(ZshCompletion())
	case "fish":
		ctx.Printer.Println(FishCompletion())
	default:
		return command.NewUsageError(fmt.Sprintf("unsupported shell %q", shell), "Supported shells: bash, zsh, fish")
	}
	return nil
}

// BashCompletion returns the bash completion script for nova.
func BashCompletion() string {
	return `# bash completion for nova
_nova_completions() {
    local cur prev commands themes
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="ls cat tree find stat du cp mv rm mkdir interactive which touch diff completion"
    themes="default minimal mono nord dracula neon cyberpunk synthwave tokyo-night catppuccin gruvbox"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        if [[ ${cur} == -* ]]; then
            COMPREPLY=( $(compgen -W "--help --version --plain --json --debug --color= --theme= --icons=" -- ${cur}) )
        else
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        fi
        return 0
    fi

    case "${prev}" in
        --theme|--theme=*)
            COMPREPLY=( $(compgen -W "${themes}" -- ${cur}) )
            return 0
            ;;
        --color|--color=*)
            COMPREPLY=( $(compgen -W "auto always never" -- ${cur}) )
            return 0
            ;;
        --icons|--icons=*)
            COMPREPLY=( $(compgen -W "auto always never" -- ${cur}) )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
    esac

    # Fallback to filesystem completion
    COMPREPLY=( $(compgen -f -- ${cur}) )
}

complete -o default -F _nova_completions nova
`
}

// ZshCompletion returns the zsh completion script for nova.
func ZshCompletion() string {
	return `#compdef nova

_nova() {
    local -a commands themes
    commands=(
        'ls:List directory contents with modern layout and colors'
        'cat:Stream and view files with syntax highlighting and paging'
        'tree:Display directory hierarchy as a visual tree'
        'find:Search files across directories by predicates'
        'stat:Display structured file status and metadata'
        'du:Estimate disk space usage with visual progress bars'
        'cp:Copy files and directories with progress'
        'mv:Move or rename files and directories safely'
        'rm:Remove files and directories with safeguards'
        'mkdir:Create directories with parent creation'
        'interactive:Interactive terminal file navigator'
        'which:Locate executables in PATH with symlink info'
        'touch:Create files or update timestamps'
        'diff:Compare files with colorized unified diff'
        'completion:Generate shell autocompletion script'
    )
    themes=(default minimal mono nord dracula neon cyberpunk synthwave tokyo-night catppuccin gruvbox)

    _arguments -C \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '(-v --version)'{-v,--version}'[Show version information]' \
        '--plain[Force unformatted plain text]' \
        '--json[Force structured JSON output]' \
        '--debug[Enable diagnostic debug logging]' \
        '--color=[Color policy]:policy:(auto always never)' \
        '--theme=[Color theme]:theme:($themes)' \
        '--icons=[Icon policy]:policy:(auto always never)' \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            _describe -t commands 'nova command' commands
            ;;
        args)
            case $words[1] in
                completion)
                    _values 'shell' bash zsh fish
                    ;;
                *)
                    _files
                    ;;
            esac
            ;;
    esac
}

_nova "$@"
`
}

// FishCompletion returns the fish completion script for nova.
func FishCompletion() string {
	return `# fish completion for nova
set -l commands ls cat tree find stat du cp mv rm mkdir interactive which touch diff completion
set -l themes default minimal mono nord dracula neon cyberpunk synthwave tokyo-night catppuccin gruvbox

complete -c nova -f

# Global flags
complete -c nova -s h -l help -d "Show help message"
complete -c nova -s v -l version -d "Show version information"
complete -c nova -l plain -d "Force unformatted plain text"
complete -c nova -l json -d "Force structured JSON output"
complete -c nova -l debug -d "Enable diagnostic debug logging"
complete -c nova -l color -x -a "auto always never" -d "Color policy"
complete -c nova -l theme -x -a "$themes" -d "Color theme"
complete -c nova -l icons -x -a "auto always never" -d "Icon policy"

# Subcommands
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a ls -d "List directory contents"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a cat -d "Stream and view files"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a tree -d "Display directory hierarchy"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a find -d "Search files across directories"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a stat -d "Display structured file status"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a du -d "Estimate disk space usage"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a cp -d "Copy files and directories"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a mv -d "Move or rename files and directories"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a rm -d "Remove files and directories"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a mkdir -d "Create directories"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a interactive -d "Interactive terminal file navigator"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a which -d "Locate executables in PATH"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a touch -d "Create files or update timestamps"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a diff -d "Compare files with colorized diff"
complete -c nova -n "not __fish_seen_subcommand_from $commands" -a completion -d "Generate shell completion script"

# Completion subcommand arguments
complete -c nova -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`
}
