package scancmd

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Rule represents a secret detection rule with regex or heuristic matcher.
type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    string // "CRITICAL", "HIGH", "MEDIUM"
	Regex       *regexp.Regexp
}

// Built-in secret detection patterns.
var DefaultRules = []Rule{
	{
		ID:          "github-pat",
		Name:        "GitHub Personal Access Token",
		Description: "Detected personal access token for GitHub",
		Severity:    "CRITICAL",
		Regex:       regexp.MustCompile(`(ghp_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59})`),
	},
	{
		ID:          "aws-key",
		Name:        "AWS Access Key ID",
		Description: "Detected AWS Access Key identifier",
		Severity:    "CRITICAL",
		Regex:       regexp.MustCompile(`(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`),
	},
	{
		ID:          "slack-webhook",
		Name:        "Slack Incoming Webhook",
		Description: "Detected Slack webhook URL with embedded credentials",
		Severity:    "HIGH",
		Regex:       regexp.MustCompile(`https://hooks\.slack\.com/services/T[a-zA-Z0-9_]+/B[a-zA-Z0-9_]+/[a-zA-Z0-9_]+`),
	},
	{
		ID:          "openai-api-key",
		Name:        "OpenAI API Key",
		Description: "Detected OpenAI service API secret key",
		Severity:    "CRITICAL",
		Regex:       regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),
	},
	{
		ID:          "stripe-key",
		Name:        "Stripe Live API Key",
		Description: "Detected Stripe production secret key",
		Severity:    "CRITICAL",
		Regex:       regexp.MustCompile(`sk_live_[0-9a-zA-Z]{24}`),
	},
	{
		ID:          "private-key",
		Name:        "Cryptographic Private Key",
		Description: "Detected unencrypted RSA/EC/SSH private key block",
		Severity:    "CRITICAL",
		Regex:       regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
	},
	{
		ID:          "generic-secret",
		Name:        "Hardcoded Secret / Password",
		Description: "Detected variable assignment with potential credential",
		Severity:    "MEDIUM",
		Regex:       regexp.MustCompile(`(?i)(api_key|apikey|secret_key|secret|password|access_token)\s*[:=]\s*["']([^"'\s]{8,})["']`),
	},
}

// CalculateShannonEntropy returns the Shannon entropy of a string (in bits per byte).
func CalculateShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]float64)
	for _, r := range s {
		freq[r]++
	}

	total := float64(len(s))
	var entropy float64
	for _, count := range freq {
		p := count / total
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// MaskSecret masks sensitive values, showing only leading and trailing context.
func MaskSecret(secret string) string {
	clean := strings.TrimSpace(secret)
	if len(clean) <= 8 {
		return "********"
	}
	headLen := 4
	tailLen := 4
	if len(clean) >= 20 {
		headLen = 6
		tailLen = 4
	}
	head := clean[:headLen]
	tail := clean[len(clean)-tailLen:]
	return fmt.Sprintf("%s****...%s", head, tail)
}
