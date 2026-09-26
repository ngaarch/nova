package syntax

import (
	"strings"
	"unicode"
)

var (
	goKeywords = map[string]bool{
		"break": true, "default": true, "func": true, "interface": true, "select": true,
		"case": true, "defer": true, "go": true, "map": true, "struct": true,
		"chan": true, "else": true, "goto": true, "package": true, "switch": true,
		"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
		"continue": true, "for": true, "import": true, "return": true, "var": true,
		"nil": true, "true": true, "false": true, "iota": true,
	}

	goTypes = map[string]bool{
		"bool": true, "byte": true, "complex64": true, "complex128": true,
		"error": true, "float32": true, "float64": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"rune": true, "string": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"any": true, "comparable": true,
	}

	pyKeywords = map[string]bool{
		"and": true, "as": true, "assert": true, "async": true, "await": true,
		"break": true, "class": true, "continue": true, "def": true, "del": true,
		"elif": true, "else": true, "except": true, "finally": true, "for": true,
		"from": true, "global": true, "if": true, "import": true, "in": true,
		"is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
		"pass": true, "raise": true, "return": true, "try": true, "while": true,
		"with": true, "yield": true, "None": true, "True": true, "False": true, "self": true,
	}

	pyTypes = map[string]bool{
		"int": true, "float": true, "str": true, "bool": true, "list": true,
		"dict": true, "set": true, "tuple": true, "bytes": true, "object": true,
	}

	jsKeywords = map[string]bool{
		"async": true, "await": true, "break": true, "case": true, "catch": true,
		"class": true, "const": true, "continue": true, "debugger": true, "default": true,
		"delete": true, "do": true, "else": true, "export": true, "extends": true,
		"finally": true, "for": true, "function": true, "if": true, "import": true,
		"in": true, "instanceof": true, "new": true, "return": true, "super": true,
		"switch": true, "this": true, "throw": true, "try": true, "typeof": true,
		"var": true, "void": true, "while": true, "with": true, "yield": true,
		"let": true, "static": true, "interface": true, "type": true, "enum": true,
		"null": true, "undefined": true, "true": true, "false": true,
	}

	jsTypes = map[string]bool{
		"string": true, "number": true, "boolean": true, "symbol": true, "any": true,
		"unknown": true, "never": true, "void": true, "Array": true, "Promise": true,
	}

	shKeywords = map[string]bool{
		"if": true, "then": true, "else": true, "elif": true, "fi": true,
		"case": true, "esac": true, "for": true, "while": true, "until": true,
		"do": true, "done": true, "in": true, "function": true, "select": true,
		"time": true, "return": true, "exit": true, "export": true, "local": true,
	}

	cKeywords = map[string]bool{
		"auto": true, "break": true, "case": true, "char": true, "const": true,
		"continue": true, "default": true, "do": true, "double": true, "else": true,
		"enum": true, "extern": true, "float": true, "for": true, "goto": true,
		"if": true, "inline": true, "int": true, "long": true, "register": true,
		"restrict": true, "return": true, "short": true, "signed": true, "sizeof": true,
		"static": true, "struct": true, "switch": true, "typedef": true, "union": true,
		"unsigned": true, "void": true, "volatile": true, "while": true,
		"fn": true, "let": true, "mut": true, "pub": true, "impl": true, "trait": true,
		"match": true, "use": true, "mod": true, "crate": true,
	}

	sqlKeywords = map[string]bool{
		"select": true, "from": true, "where": true, "insert": true, "into": true,
		"update": true, "delete": true, "join": true, "inner": true, "left": true,
		"right": true, "outer": true, "on": true, "group": true, "by": true,
		"order": true, "having": true, "limit": true, "create": true, "table": true,
		"drop": true, "alter": true, "index": true, "primary": true, "key": true,
		"foreign": true, "references": true, "values": true, "distinct": true,
		"union": true, "all": true, "as": true, "and": true, "or": true, "not": true,
		"null": true, "is": true, "in": true, "between": true, "like": true,
		"case": true, "when": true, "then": true, "else": true, "end": true,
	}
)

func getKeywordsAndTypes(lang string) (map[string]bool, map[string]bool) {
	switch strings.ToLower(lang) {
	case "go":
		return goKeywords, goTypes
	case "python", "py":
		return pyKeywords, pyTypes
	case "javascript", "js", "typescript", "ts":
		return jsKeywords, jsTypes
	case "bash", "sh", "zsh":
		return shKeywords, nil
	case "c", "cpp", "cxx", "rust", "rs":
		return cKeywords, nil
	case "sql":
		return sqlKeywords, nil
	default:
		return goKeywords, goTypes
	}
}

func tokenizeCode(line string, lang string, state *State) []Token {
	keywords, types := getKeywordsAndTypes(lang)
	var tokens []Token
	runes := []rune(line)
	n := len(runes)
	i := 0

	if state.InBlockComment {
		endIdx := strings.Index(line, state.BlockCommentEnd)
		if endIdx != -1 {
			commentLen := len([]rune(line[:endIdx])) + len([]rune(state.BlockCommentEnd))
			tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[:commentLen])})
			state.InBlockComment = false
			state.BlockCommentEnd = ""
			i = commentLen
		} else {
			tokens = append(tokens, Token{Type: TokenComment, Value: line})
			return tokens
		}
	}

	if state.InRawString {
		endIdx := strings.Index(line, state.RawStringDelim)
		if endIdx != -1 {
			strLen := len([]rune(line[:endIdx])) + len([]rune(state.RawStringDelim))
			tokens = append(tokens, Token{Type: TokenString, Value: string(runes[:strLen])})
			state.InRawString = false
			state.RawStringDelim = ""
			i = strLen
		} else {
			tokens = append(tokens, Token{Type: TokenString, Value: line})
			return tokens
		}
	}

	for i < n {
		r := runes[i]

		if unicode.IsSpace(r) {
			start := i
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenText, Value: string(runes[start:i])})
			continue
		}

		isHashComment := (r == '#' && (lang == "python" || lang == "py" || lang == "bash" || lang == "sh" || lang == "ruby" || lang == "yaml"))
		isDashComment := (r == '-' && i+1 < n && runes[i+1] == '-' && lang == "sql")
		isSlashComment := (r == '/' && i+1 < n && runes[i+1] == '/')

		if isHashComment || isDashComment || isSlashComment {
			tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[i:])})
			break
		}

		if r == '/' && i+1 < n && runes[i+1] == '*' {
			start := i
			i += 2
			closed := false
			for i+1 < n {
				if runes[i] == '*' && runes[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				i++
			}
			if closed {
				tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[start:i])})
			} else {
				state.InBlockComment = true
				state.BlockCommentEnd = "*/"
				tokens = append(tokens, Token{Type: TokenComment, Value: string(runes[start:])})
				break
			}
			continue
		}

		if r == '`' {
			start := i
			i++
			closed := false
			for i < n {
				if runes[i] == '`' {
					i++
					closed = true
					break
				}
				i++
			}
			if closed {
				tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
			} else {
				state.InRawString = true
				state.RawStringDelim = "`"
				tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:])})
				break
			}
			continue
		}

		if (r == '"' || r == '\'') && i+2 < n && runes[i+1] == r && runes[i+2] == r && (lang == "python" || lang == "py") {
			delim := string([]rune{r, r, r})
			start := i
			i += 3
			closed := false
			for i+2 < n {
				if runes[i] == r && runes[i+1] == r && runes[i+2] == r {
					i += 3
					closed = true
					break
				}
				i++
			}
			if closed {
				tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
			} else {
				state.InRawString = true
				state.RawStringDelim = delim
				tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:])})
				break
			}
			continue
		}

		if r == '"' || r == '\'' {
			quote := r
			start := i
			i++
			escaped := false
			for i < n {
				if escaped {
					escaped = false
				} else if runes[i] == '\\' {
					escaped = true
				} else if runes[i] == quote {
					i++
					break
				}
				i++
			}
			tokens = append(tokens, Token{Type: TokenString, Value: string(runes[start:i])})
			continue
		}

		if r == '$' && (lang == "bash" || lang == "sh" || lang == "zsh") {
			start := i
			i++
			if i < n && runes[i] == '{' {
				for i < n && runes[i] != '}' {
					i++
				}
				if i < n && runes[i] == '}' {
					i++
				}
			} else {
				for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
					i++
				}
			}
			tokens = append(tokens, Token{Type: TokenAttr, Value: string(runes[start:i])})
			continue
		}

		if unicode.IsDigit(r) || (r == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			i++
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.' || runes[i] == 'x' || runes[i] == 'X' ||
				(runes[i] >= 'a' && runes[i] <= 'f') || (runes[i] >= 'A' && runes[i] <= 'F') ||
				runes[i] == '_' || runes[i] == 'e' || runes[i] == 'E') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: string(runes[start:i])})
			continue
		}

		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			val := string(runes[start:i])
			lowerVal := strings.ToLower(val)

			if keywords != nil && (keywords[val] || (lang == "sql" && keywords[lowerVal])) {
				tokens = append(tokens, Token{Type: TokenKeyword, Value: val})
			} else if types != nil && types[val] {
				tokens = append(tokens, Token{Type: TokenTypeIdent, Value: val})
			} else {
				tokens = append(tokens, Token{Type: TokenText, Value: val})
			}
			continue
		}

		if strings.ContainsRune(":=<>!+-*/%&|^~?", r) {
			start := i
			for i < n && strings.ContainsRune(":=<>!+-*/%&|^~?", runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenOperator, Value: string(runes[start:i])})
			continue
		}

		if strings.ContainsRune("()[]{},;.", r) {
			tokens = append(tokens, Token{Type: TokenPunctuation, Value: string(r)})
			i++
			continue
		}

		tokens = append(tokens, Token{Type: TokenText, Value: string(r)})
		i++
	}

	return tokens
}
