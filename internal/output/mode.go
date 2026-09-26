package output

// Mode determines the format and styling of command output.
type Mode string

const (
	// ModeHuman produces rich, responsive, colored output designed for interactive terminals.
	ModeHuman Mode = "human"
	// ModePlain produces clean, uncolored, deterministic text for pipelines and non-interactive use.
	ModePlain Mode = "plain"
	// ModeJSON produces structured, valid JSON documents for programmatic ingestion.
	ModeJSON Mode = "json"
)

// String returns the string representation of the output mode.
func (m Mode) String() string {
	return string(m)
}

// ResolveMode determines the effective output mode given explicit flags and TTY detection.
// Follows the rule: JSON > Plain > Auto (TTY ? Human : Plain).
func ResolveMode(flagPlain, flagJSON bool, isStdoutTTY bool) Mode {
	if flagJSON {
		return ModeJSON
	}
	if flagPlain {
		return ModePlain
	}
	if !isStdoutTTY {
		return ModePlain
	}
	return ModeHuman
}
