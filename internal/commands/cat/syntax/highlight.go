package syntax

import (
	"strings"

	"nova/internal/terminal"
	"nova/internal/theme"
)

// HighlightToken returns the ANSI-formatted string for a single token.
func HighlightToken(tok Token, th *theme.Theme, profile terminal.ColorProfile) string {
	if profile == terminal.ColorNone || th == nil {
		return tok.Value
	}

	var role theme.Role
	switch tok.Type {
	case TokenKeyword:
		role = theme.RoleAccent
	case TokenTypeIdent:
		role = theme.RoleExecutable
	case TokenString:
		role = theme.RoleSuccess
	case TokenNumber:
		role = theme.RoleWarning
	case TokenComment:
		role = theme.RoleMuted
	case TokenOperator:
		role = theme.RoleMuted
	case TokenPunctuation:
		role = theme.RoleMuted
	case TokenTag:
		role = theme.RoleAccent
	case TokenAttr:
		role = theme.RoleExecutable
	case TokenHeading:
		role = theme.RoleAccent
	case TokenListMarker:
		role = theme.RoleAccent
	case TokenBlockQuote:
		role = theme.RoleMuted
	default:
		return tok.Value
	}

	return th.Format(role, tok.Value, profile)
}

// HighlightLine tokenizes and formats a single line with theme colors.
func HighlightLine(line string, lang string, state *State, th *theme.Theme, profile terminal.ColorProfile) string {
	if profile == terminal.ColorNone || lang == "text" || lang == "" {
		return line
	}

	tokens := Tokenize(line, lang, state)
	var b strings.Builder
	for _, tok := range tokens {
		b.WriteString(HighlightToken(tok, th, profile))
	}
	return b.String()
}
