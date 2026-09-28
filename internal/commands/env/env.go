package env

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// Variable represents one parsed environment variable.
type Variable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DisplayValue string `json:"display_value"`
	Category     string `json:"category"`
	IsSecret     bool   `json:"is_secret"`
}

// Result holds the variables list and audit summary.
type Result struct {
	Variables    []Variable     `json:"variables"`
	Total        int            `json:"total"`
	SecretsCount int            `json:"secrets_count"`
	Categories   map[string]int `json:"categories"`
}

// Command returns the registered Command instance for env.
func Command() *command.Command {
	return &command.Command{
		Name:        "env",
		Aliases:     []string{"environ", "envinfo"},
		Summary:     "Audit environment variables with automatic secret masking, categorization, and export formats",
		Usage:       "nova env [flags] [filter]",
		Description: "Inspect process environment variables, safely mask sensitive credentials, and generate shell exports.",
		Phase:       18,
		Run:         Run,
	}
}

// Run executes the env command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	envVars := os.Environ()

	result := inspectEnvironment(envVars, opts)
	return Render(ctx, result, opts)
}

func inspectEnvironment(rawEnv []string, opts Options) Result {
	var filterRe *regexp.Regexp
	if opts.Filter != "" {
		filterRe, _ = regexp.Compile("(?i)" + regexp.QuoteMeta(opts.Filter))
	}

	var vars []Variable
	categoryCounts := make(map[string]int)
	secretsCount := 0

	for _, entry := range rawEnv {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key, val := parts[0], parts[1]

		if filterRe != nil && !filterRe.MatchString(key) {
			continue
		}

		isSecret := isSecretKey(key)
		if isSecret {
			secretsCount++
		}

		displayVal := val
		if isSecret && !opts.ShowSecrets {
			displayVal = "********"
		}

		cat := categorizeKey(key)
		categoryCounts[cat]++

		vars = append(vars, Variable{
			Key:          key,
			Value:        val,
			DisplayValue: displayVal,
			Category:     cat,
			IsSecret:     isSecret,
		})
	}

	// Sort alphabetically by Key
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Key < vars[j].Key
	})

	return Result{
		Variables:    vars,
		Total:        len(vars),
		SecretsCount: secretsCount,
		Categories:   categoryCounts,
	}
}

func isSecretKey(key string) bool {
	upper := strings.ToUpper(key)
	secretIndicators := []string{
		"SECRET", "TOKEN", "PASSWORD", "PASSWD", "KEY",
		"AUTH", "PRIVATE", "CREDENTIAL", "ACCESS_KEY",
		"APIKEY", "SIG", "SESSION", "BEARER",
	}

	// Exceptions that contain "KEY" or "PATH" but aren't secrets
	if upper == "KEYMAP" || upper == "KEYBOARD" || upper == "COLORKEY" {
		return false
	}

	for _, ind := range secretIndicators {
		if strings.Contains(upper, ind) {
			return true
		}
	}
	return false
}

func categorizeKey(key string) string {
	upper := strings.ToUpper(key)

	// Development Runtimes
	runtimePrefixes := []string{"GO", "NODE", "NPM", "PYTHON", "PY", "RUST", "CARGO", "JAVA", "RUBY", "PHP", "DENO", "BUN", "DOTNET", "VIRTUAL_ENV"}
	for _, p := range runtimePrefixes {
		if strings.HasPrefix(upper, p) {
			return "Development Runtimes"
		}
	}

	// Cloud & DevOps
	cloudPrefixes := []string{"AWS", "DOCKER", "KUBE", "CONTAINER", "CI", "GITHUB", "GITLAB", "HEROKU", "VERCEL"}
	for _, p := range cloudPrefixes {
		if strings.HasPrefix(upper, p) {
			return "Cloud & DevOps"
		}
	}

	// System & Shell
	systemKeys := []string{"PATH", "SHELL", "USER", "LOGNAME", "HOME", "TERM", "LANG", "LC_", "EDITOR", "VISUAL", "PAGER", "PWD", "OLDPWD", "SHLVL", "HOSTNAME"}
	for _, s := range systemKeys {
		if strings.HasPrefix(upper, s) {
			return "System & Shell"
		}
	}

	return "General"
}
