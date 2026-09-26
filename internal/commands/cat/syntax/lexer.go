package syntax

import (
	"strings"
	"unicode"
)

// Tokenize processes a single line of text for the specified language and returns semantic tokens.
func Tokenize(line string, lang string, state *State) []Token {
	if state == nil {
		state = &State{}
	}

	// Format-specific handlers
	switch strings.ToLower(lang) {
	case "markdown", "md":
		return tokenizeMarkdown(line, state)
	case "json":
		return tokenizeJSON(line, state)
	case "yaml", "yml":
		return tokenizeYAML(line, state)
	case "html", "xml", "svg":
		return tokenizeHTML(line, state)
	default:
		return tokenizeCode(line, lang, state)
	}
}

// tokenizeMarkdown parses Markdown structures and nested code fences.
func tokenizeMarkdown(line string, state *State) []Token {
	trimmed := strings.TrimSpace(line)

	// Code fence toggle
	if strings.HasPrefix(trimmed, "```") {
		if state.InCodeFence {
			state.InCodeFence = false
			state.CodeFenceLang = ""
			return []Token{{Type: TokenComment, Value: line}}
		}
		state.InCodeFence = true
		lang := strings.TrimPrefix(trimmed, "```")
		state.CodeFenceLang = strings.TrimSpace(lang)
		return []Token{{Type: TokenComment, Value: line}}
	}

	// Inside code fence: delegate to the code lexer
	if state.InCodeFence {
		lang := state.CodeFenceLang
		if lang == "" {
			lang = "text"
		}
		return tokenizeCode(line, lang, state)
	}

	// Headings
	if strings.HasPrefix(trimmed, "#") {
		count := 0
		for _, r := range trimmed {
			if r == '#' {
				count++
			} else {
				break
			}
		}
		if count >= 1 && count <= 6 && (len(trimmed) == count || trimmed[count] == ' ') {
			return []Token{{Type: TokenHeading, Value: line}}
		}
	}

	// Blockquote
	if strings.HasPrefix(trimmed, ">") {
		idx := strings.Index(line, ">")
		prefix := line[:idx]
		quoteMarker := line[idx : idx+1]
		rest := line[idx+1:]
		var tokens []Token
		if prefix != "" {
			tokens = append(tokens, Token{Type: TokenText, Value: prefix})
		}
		tokens = append(tokens, Token{Type: TokenBlockQuote, Value: quoteMarker})
		if rest != "" {
			tokens = append(tokens, Token{Type: TokenText, Value: rest})
		}
		return tokens
	}

	// List markers
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		idx := strings.IndexAny(line, "-*+")
		prefix := line[:idx]
		marker := line[idx : idx+2]
		rest := line[idx+2:]
		var tokens []Token
		if prefix != "" {
			tokens = append(tokens, Token{Type: TokenText, Value: prefix})
		}
		tokens = append(tokens, Token{Type: TokenListMarker, Value: marker})
		if rest != "" {
			tokens = append(tokens, tokenizeInlineMarkdown(rest)...)
		}
		return tokens
	}

	// Default inline markdown parsing
	return tokenizeInlineMarkdown(line)
}

func tokenizeInlineMarkdown(line string) []Token {
	var tokens []Token
	runes := []rune(line)
	n := len(runes)
	i := 0

	for i < n {
		// Inline backtick code
		if runes[i] == '`' {
			start := i
			i++
			for i < n && runes[i] != '`' {
				i++
			}
			if i < n && runes[i] == '`' {
				i++
			}
			tokens = append(tokens, Token{Type: TokenKeyword, Value: string(runes[start:i])})
			continue
		}

		// Text accumulator
		start := i
		for i < n && runes[i] != '`' {
			i++
		}
		if i > start {
			tokens = append(tokens, Token{Type: TokenText, Value: string(runes[start:i])})
		}
	}

	return tokens
}

// tokenizeJSON handles JSON syntax (keys vs values, numbers, booleans, null).
func tokenizeJSON(line string, state *State) []Token {
	var tokens []Token
	runes := []rune(line)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]

		// Whitespace
		if unicode.IsSpace(r) {
			start := i
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenText, Value: string(runes[start:i])})
			continue
		}

		// String
		if r == '"' {
			start := i
			i++
			escaped := false
			for i < n {
				if escaped {
					escaped = false
				} else if runes[i] == '\\' {
					escaped = true
				} else if runes[i] == '"' {
					i++
					break
				}
				i++
			}
			strVal := string(runes[start:i])

			// Check if this string is followed by a colon (making it a JSON key)
			isKey := false
			j := i
			for j < n && unicode.IsSpace(runes[j]) {
				j++
			}
			if j < n && runes[j] == ':' {
				isKey = true
			}

			if isKey {
				tokens = append(tokens, Token{Type: TokenAttr, Value: strVal})
			} else {
				tokens = append(tokens, Token{Type: TokenString, Value: strVal})
			}
			continue
		}

		// Numbers
		if r == '-' || (r >= '0' && r <= '9') {
			start := i
			i++
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.' || runes[i] == 'e' || runes[i] == 'E' || runes[i] == '+' || runes[i] == '-') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: string(runes[start:i])})
			continue
		}

		// Punctuation & Structural tokens
		if r == '{' || r == '}' || r == '[' || r == ']' || r == ':' || r == ',' {
			tokens = append(tokens, Token{Type: TokenPunctuation, Value: string(r)})
			i++
			continue
		}

		// Identifiers (true, false, null)
		if unicode.IsLetter(r) {
			start := i
			for i < n && unicode.IsLetter(runes[i]) {
				i++
			}
			val := string(runes[start:i])
			if val == "true" || val == "false" || val == "null" {
				tokens = append(tokens, Token{Type: TokenKeyword, Value: val})
			} else {
				tokens = append(tokens, Token{Type: TokenText, Value: val})
			}
			continue
		}

		// Other characters
		tokens = append(tokens, Token{Type: TokenText, Value: string(r)})
		i++
	}

	return tokens
}

// tokenizeYAML parses YAML lines (keys, values, comments).
func tokenizeYAML(line string, state *State) []Token {
	var tokens []Token
	runes := []rune(line)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]

		// Whitespace
		if unicode.IsSpace(r) {
			start := i
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenText, Value: string(runes[start:i])})
			continue
		}

		// Comment
		if r == '#' {
			tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[i:])})
			break
		}

		// String
		if r == '"' || r == '\'' {
			quote := r
			start := i
			i++
			for i < n && runes[i] != quote {
				if runes[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			if i < n && runes[i] == quote {
				i++
			}
			tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
			continue
		}

		// Key: word followed by colon
		start := i
		for i < n && runes[i] != ':' && runes[i] != '#' && runes[i] != '\n' {
			i++
		}
		if i < n && runes[i] == ':' {
			tokens = append(tokens, Token{Type: TokenAttr, Value: string(runes[start:i])})
			tokens = append(tokens, Token{Type: TokenPunctuation, Value: ":"})
			i++
			continue
		}

		// If no colon found, reset and consume token
		if i > start {
			val := string(runes[start:i])
			trimmedVal := strings.TrimSpace(val)
			if trimmedVal == "true" || trimmedVal == "false" || trimmedVal == "null" || trimmedVal == "yes" || trimmedVal == "no" {
				tokens = append(tokens, Token{Type: TokenKeyword, Value: val})
			} else {
				tokens = append(tokens, Token{Type: TokenText, Value: val})
			}
			continue
		}

		tokens = append(tokens, Token{Type: TokenText, Value: string(r)})
		i++
	}

	return tokens
}

// tokenizeHTML parses HTML/XML tags and attributes.
func tokenizeHTML(line string, state *State) []Token {
	var tokens []Token
	runes := []rune(line)
	n := len(runes)
	i := 0

	for i < n {
		// Comments <!-- ... -->
		if i+3 < n && string(runes[i:i+4]) == "<!--" {
			start := i
			i += 4
			for i+2 < n && string(runes[i:i+3]) != "-->" {
				i++
			}
			if i+2 < n {
				i += 3
			} else {
				i = n
			}
			tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[start:i])})
			continue
		}

		// Tag
		if runes[i] == '<' {
			start := i
			i++
			if i < n && runes[i] == '/' {
				i++
			}
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '-' || runes[i] == ':') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenTag, Value: string(runes[start:i])})
			continue
		}

		// Closing tag bracket
		if runes[i] == '>' {
			tokens = append(tokens, Token{Type: TokenTag, Value: ">"})
			i++
			continue
		}

		// String attribute value
		if runes[i] == '"' || runes[i] == '\'' {
			quote := runes[i]
			start := i
			i++
			for i < n && runes[i] != quote {
				i++
			}
			if i < n && runes[i] == quote {
				i++
			}
			tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
			continue
		}

		// Attribute name
		if unicode.IsLetter(runes[i]) {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '-' || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenAttr, Value: string(runes[start:i])})
			continue
		}

		tokens = append(tokens, Token{Type: TokenText, Value: string(runes[i])})
		i++
	}

	return tokens
}
