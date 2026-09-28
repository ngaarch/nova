package calccmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/renderer"
)

// CalcResult holds evaluation outputs across bases and representations.
type CalcResult struct {
	Expression string  `json:"expression"`
	Value      float64 `json:"value"`
	Formatted  string  `json:"formatted"`
	IsInteger  bool    `json:"is_integer"`
	Hex        string  `json:"hex,omitempty"`
	Binary     string  `json:"binary,omitempty"`
	Octal      string  `json:"octal,omitempty"`
	HumanBytes string  `json:"human_bytes,omitempty"`
}

// Command returns the registered Command instance for calc.
func Command() *command.Command {
	return &command.Command{
		Name:        "calc",
		Aliases:     []string{"expr", "math", "eval"},
		Summary:     "Scientific, programmer, and byte-unit terminal calculator",
		Usage:       "nova calc [expression] [flags]",
		Description: "Evaluate arithmetic, algebraic, bitwise expressions, and multi-base number conversions.",
		Phase:       24,
		Run:         Run,
	}
}

// Run executes the calc command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	if strings.TrimSpace(opts.Expression) == "" {
		return fmt.Errorf("expression required: e.g. nova calc \"2^16 + 1024 * 8\"")
	}

	val, err := Evaluate(opts.Expression)
	if err != nil {
		return fmt.Errorf("calc: %w", err)
	}

	res := FormatResult(opts.Expression, val, opts)

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderPlain(ctx.Stdout, res, opts)
	} else {
		RenderDashboard(ctx.Stdout, res, ctx)
	}
	return nil
}

// FormatResult converts a float value into CalcResult with base breakdown.
func FormatResult(expr string, val float64, opts Options) *CalcResult {
	res := &CalcResult{
		Expression: expr,
		Value:      val,
	}

	isWhole := val == math.Trunc(val) && !math.IsNaN(val) && !math.IsInf(val, 0)
	res.IsInteger = isWhole

	if isWhole && val >= float64(math.MinInt64) && val <= float64(math.MaxInt64) {
		intVal := int64(val)
		res.Formatted = strconv.FormatInt(intVal, 10)
		res.Hex = fmt.Sprintf("0x%X", uint64(intVal))
		res.Binary = fmt.Sprintf("0b%b", uint64(intVal))
		res.Octal = fmt.Sprintf("0o%o", uint64(intVal))
		if intVal > 0 {
			res.HumanBytes = renderer.FormatSize(intVal, true)
		}
	} else {
		prec := opts.Precision
		if prec < 0 {
			prec = 6
		}
		res.Formatted = strconv.FormatFloat(val, 'f', prec, 64)
		res.Formatted = strings.TrimRight(strings.TrimRight(res.Formatted, "0"), ".")
	}

	return res
}

// Evaluate parses and computes mathematical expressions.
func Evaluate(expr string) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	p := &parser{tokens: tokens, pos: 0}
	val, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.tokens) {
		return 0, fmt.Errorf("unexpected token '%s' at position %d", p.tokens[p.pos].val, p.pos)
	}
	return val, nil
}

type tokenType int

const (
	tokNum tokenType = iota
	tokIdent
	tokOp
	tokLParen
	tokRParen
	tokComma
)

type token struct {
	typ tokenType
	val string
	num float64
}

func tokenize(s string) ([]token, error) {
	var tokens []token
	runes := []rune(s)
	i := 0
	n := len(runes)

	for i < n {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		if r == '(' {
			tokens = append(tokens, token{typ: tokLParen, val: "("})
			i++
			continue
		}
		if r == ')' {
			tokens = append(tokens, token{typ: tokRParen, val: ")"})
			i++
			continue
		}
		if r == ',' {
			tokens = append(tokens, token{typ: tokComma, val: ","})
			i++
			continue
		}

		// Numbers: decimals, 0x hex, 0b bin, 0o oct
		if unicode.IsDigit(r) || (r == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			isHex := false
			isBin := false
			isOct := false

			if r == '0' && i+1 < n {
				switch runes[i+1] {
				case 'x', 'X':
					isHex = true
					i += 2
				case 'b', 'B':
					isBin = true
					i += 2
				case 'o', 'O':
					isOct = true
					i += 2
				}
			}

			if isHex {
				for i < n && (unicode.IsDigit(runes[i]) || (runes[i] >= 'a' && runes[i] <= 'f') || (runes[i] >= 'A' && runes[i] <= 'F')) {
					i++
				}
				numStr := string(runes[start+2 : i])
				val, err := strconv.ParseInt(numStr, 16, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid hex number: %s", string(runes[start:i]))
				}
				tokens = append(tokens, token{typ: tokNum, val: string(runes[start:i]), num: float64(val)})
				continue
			}

			if isBin {
				for i < n && (runes[i] == '0' || runes[i] == '1') {
					i++
				}
				numStr := string(runes[start+2 : i])
				val, err := strconv.ParseInt(numStr, 2, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid binary number: %s", string(runes[start:i]))
				}
				tokens = append(tokens, token{typ: tokNum, val: string(runes[start:i]), num: float64(val)})
				continue
			}

			if isOct {
				for i < n && (runes[i] >= '0' && runes[i] <= '7') {
					i++
				}
				numStr := string(runes[start+2 : i])
				val, err := strconv.ParseInt(numStr, 8, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid octal number: %s", string(runes[start:i]))
				}
				tokens = append(tokens, token{typ: tokNum, val: string(runes[start:i]), num: float64(val)})
				continue
			}

			// Standard decimal
			hasDot := false
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				if runes[i] == '.' {
					if hasDot {
						break
					}
					hasDot = true
				}
				i++
			}
			numStr := string(runes[start:i])
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", numStr)
			}

			// Check for byte unit suffixes like KB, MB, GB
			unitMult := 1.0
			unitLen := 0
			if i < n && unicode.IsLetter(runes[i]) {
				uStart := i
				for i < n && unicode.IsLetter(runes[i]) {
					i++
				}
				uStr := strings.ToUpper(string(runes[uStart:i]))
				switch uStr {
				case "B":
					unitMult = 1
				case "K", "KB", "KIB":
					unitMult = 1024
				case "M", "MB", "MIB":
					unitMult = 1024 * 1024
				case "G", "GB", "GIB":
					unitMult = 1024 * 1024 * 1024
				case "T", "TB", "TIB":
					unitMult = 1024 * 1024 * 1024 * 1024
				case "P", "PB", "PIB":
					unitMult = 1024 * 1024 * 1024 * 1024 * 1024
				default:
					i = uStart // not a unit
				}
				if unitMult > 1 {
					unitLen = i - uStart
				}
			}
			_ = unitLen

			tokens = append(tokens, token{typ: tokNum, val: numStr, num: val * unitMult})
			continue
		}

		// Identifiers (functions or constants)
		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			name := string(runes[start:i])
			tokens = append(tokens, token{typ: tokIdent, val: name})
			continue
		}

		// Multi-character operators like **, <<, >>
		if i+1 < n {
			two := string(runes[i : i+2])
			if two == "**" || two == "<<" || two == ">>" {
				tokens = append(tokens, token{typ: tokOp, val: two})
				i += 2
				continue
			}
		}

		// Single-character operators
		if strings.ContainsRune("+-*/%^&|", r) {
			tokens = append(tokens, token{typ: tokOp, val: string(r)})
			i++
			continue
		}

		return nil, fmt.Errorf("unexpected character '%c'", r)
	}

	return tokens, nil
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) parseExpr() (float64, error) {
	return p.parseBitwise()
}

func (p *parser) parseBitwise() (float64, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokOp && (tok.val == "&" || tok.val == "|" || tok.val == "<<" || tok.val == ">>") {
			p.pos++
			right, err := p.parseAddSub()
			if err != nil {
				return 0, err
			}
			iLeft := int64(left)
			iRight := int64(right)
			switch tok.val {
			case "&":
				left = float64(iLeft & iRight)
			case "|":
				left = float64(iLeft | iRight)
			case "<<":
				left = float64(iLeft << uint64(iRight))
			case ">>":
				left = float64(iLeft >> uint64(iRight))
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseAddSub() (float64, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokOp && (tok.val == "+" || tok.val == "-") {
			p.pos++
			right, err := p.parseMulDiv()
			if err != nil {
				return 0, err
			}
			if tok.val == "+" {
				left += right
			} else {
				left -= right
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseMulDiv() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokOp && (tok.val == "*" || tok.val == "/" || tok.val == "%") {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if tok.val == "*" {
				left *= right
			} else if tok.val == "/" {
				if right == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				left /= right
			} else if tok.val == "%" {
				if right == 0 {
					return 0, fmt.Errorf("modulo by zero")
				}
				left = math.Mod(left, right)
			}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parsePower() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokOp && (tok.val == "^" || tok.val == "**") {
			p.pos++
			right, err := p.parsePower() // right-associative
			if err != nil {
				return 0, err
			}
			left = math.Pow(left, right)
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseUnary() (float64, error) {
	if p.pos >= len(p.tokens) {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	tok := p.tokens[p.pos]
	if tok.typ == tokOp && (tok.val == "+" || tok.val == "-") {
		p.pos++
		val, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		if tok.val == "-" {
			return -val, nil
		}
		return val, nil
	}

	return p.parsePrimary()
}

func (p *parser) parsePrimary() (float64, error) {
	if p.pos >= len(p.tokens) {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	tok := p.tokens[p.pos]
	p.pos++

	if tok.typ == tokNum {
		return tok.num, nil
	}

	if tok.typ == tokLParen {
		val, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos].typ != tokRParen {
			return 0, fmt.Errorf("missing closing parenthesis ')'")
		}
		p.pos++
		return val, nil
	}

	if tok.typ == tokIdent {
		name := strings.ToLower(tok.val)

		// Constants
		switch name {
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		case "tau":
			return 2 * math.Pi, nil
		case "phi":
			return 1.618033988749895, nil
		}

		// Function call: ident ( args... )
		if p.pos < len(p.tokens) && p.tokens[p.pos].typ == tokLParen {
			p.pos++ // consume '('
			arg, err := p.parseExpr()
			if err != nil {
				return 0, err
			}
			if p.pos >= len(p.tokens) || p.tokens[p.pos].typ != tokRParen {
				return 0, fmt.Errorf("missing closing parenthesis for function %s", name)
			}
			p.pos++ // consume ')'

			switch name {
			case "sqrt":
				if arg < 0 {
					return 0, fmt.Errorf("sqrt of negative number")
				}
				return math.Sqrt(arg), nil
			case "cbrt":
				return math.Cbrt(arg), nil
			case "sin":
				return math.Sin(arg), nil
			case "cos":
				return math.Cos(arg), nil
			case "tan":
				return math.Tan(arg), nil
			case "abs":
				return math.Abs(arg), nil
			case "floor":
				return math.Floor(arg), nil
			case "ceil":
				return math.Ceil(arg), nil
			case "round":
				return math.Round(arg), nil
			case "log", "ln":
				if arg <= 0 {
					return 0, fmt.Errorf("log of non-positive number")
				}
				return math.Log(arg), nil
			case "log2":
				if arg <= 0 {
					return 0, fmt.Errorf("log2 of non-positive number")
				}
				return math.Log2(arg), nil
			case "log10":
				if arg <= 0 {
					return 0, fmt.Errorf("log10 of non-positive number")
				}
				return math.Log10(arg), nil
			case "exp":
				return math.Exp(arg), nil
			default:
				return 0, fmt.Errorf("unknown function '%s'", name)
			}
		}

		return 0, fmt.Errorf("unknown identifier '%s'", name)
	}

	return 0, fmt.Errorf("unexpected token '%s'", tok.val)
}
