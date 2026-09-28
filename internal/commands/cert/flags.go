package certcmd

import (
	"strings"
	"time"
)

// Mode defines whether inspecting a TLS certificate or decoding a JWT.
type Mode string

const (
	ModeTLS Mode = "tls"
	ModeJWT Mode = "jwt"
)

// Options holds configuration for cert command.
type Options struct {
	Mode      Mode          `json:"mode"`
	Target    string        `json:"target"`
	Timeout   time.Duration `json:"timeout"`
	Insecure  bool          `json:"insecure"`
	Plain     bool          `json:"plain"`
	JSON      bool          `json:"json"`
}

// ParseFlags parses command line arguments for the cert command.
func ParseFlags(args []string) Options {
	opts := Options{
		Mode:    ModeTLS,
		Timeout: 5 * time.Second,
	}

	var positional []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
			i++
		} else if arg == "--json" {
			opts.JSON = true
			i++
		} else if arg == "-k" || arg == "--insecure" {
			opts.Insecure = true
			i++
		} else if strings.HasPrefix(arg, "--timeout=") {
			if d, err := time.ParseDuration(strings.TrimPrefix(arg, "--timeout=")); err == nil {
				opts.Timeout = d
			}
			i++
		} else if (arg == "--timeout" || arg == "-t") && i+1 < len(args) {
			if d, err := time.ParseDuration(args[i+1]); err == nil {
				opts.Timeout = d
			}
			i += 2
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
		} else {
			i++
		}
	}

	if len(positional) > 0 {
		first := strings.ToLower(positional[0])
		if first == "jwt" || first == "token" {
			opts.Mode = ModeJWT
			if len(positional) > 1 {
				opts.Target = positional[1]
			}
		} else if first == "inspect" || first == "tls" || first == "ssl" || first == "check" {
			opts.Mode = ModeTLS
			if len(positional) > 1 {
				opts.Target = positional[1]
			}
		} else {
			// If target contains 2 dots (like eyJhbGci...eyJzdWI...) it's a JWT token
			if strings.Count(positional[0], ".") == 2 && strings.HasPrefix(positional[0], "ey") {
				opts.Mode = ModeJWT
				opts.Target = positional[0]
			} else {
				opts.Mode = ModeTLS
				opts.Target = positional[0]
			}
		}
	}

	return opts
}
