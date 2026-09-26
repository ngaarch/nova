package syntax

import (
	"strings"
	"testing"

	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestTokenizeGo(t *testing.T) {
	state := &State{}
	tokens := Tokenize("func main() { var x int = 42 // answer", "go", state)

	hasFunc := false
	hasInt := false
	has42 := false
	hasComment := false

	for _, tok := range tokens {
		if tok.Type == TokenKeyword && tok.Value == "func" {
			hasFunc = true
		}
		if tok.Type == TokenTypeIdent && tok.Value == "int" {
			hasInt = true
		}
		if tok.Type == TokenNumber && tok.Value == "42" {
			has42 = true
		}
		if tok.Type == TokenComment && strings.Contains(tok.Value, "answer") {
			hasComment = true
		}
	}

	if !hasFunc {
		t.Errorf("expected func keyword, got %+v", tokens)
	}
	if !hasInt {
		t.Errorf("expected int type, got %+v", tokens)
	}
	if !has42 {
		t.Errorf("expected 42 number, got %+v", tokens)
	}
	if !hasComment {
		t.Errorf("expected comment, got %+v", tokens)
	}
}

func TestTokenizeJSON(t *testing.T) {
	state := &State{}
	tokens := Tokenize("{\"name\": \"nova\", \"version\": 1, \"active\": true}", "json", state)

	hasKey := false
	hasVal := false
	hasNum := false
	hasBool := false

	for _, tok := range tokens {
		if tok.Type == TokenAttr && tok.Value == "\"name\"" {
			hasKey = true
		}
		if tok.Type == TokenString && tok.Value == "\"nova\"" {
			hasVal = true
		}
		if tok.Type == TokenNumber && tok.Value == "1" {
			hasNum = true
		}
		if tok.Type == TokenKeyword && tok.Value == "true" {
			hasBool = true
		}
	}

	if !hasKey {
		t.Errorf("expected JSON key attribute, got %+v", tokens)
	}
	if !hasVal {
		t.Errorf("expected JSON string value, got %+v", tokens)
	}
	if !hasNum {
		t.Errorf("expected JSON number, got %+v", tokens)
	}
	if !hasBool {
		t.Errorf("expected JSON boolean, got %+v", tokens)
	}
}

func TestTokenizeMarkdown(t *testing.T) {
	state := &State{}
	hTokens := Tokenize("# Heading 1", "markdown", state)
	if len(hTokens) == 0 || hTokens[0].Type != TokenHeading {
		t.Errorf("expected TokenHeading, got %+v", hTokens)
	}

	lTokens := Tokenize("- List item", "markdown", state)
	if len(lTokens) < 2 || lTokens[0].Type != TokenListMarker {
		t.Errorf("expected TokenListMarker, got %+v", lTokens)
	}

	bTokens := Tokenize("> Blockquote", "markdown", state)
	if len(bTokens) < 2 || bTokens[0].Type != TokenBlockQuote {
		t.Errorf("expected TokenBlockQuote, got %+v", bTokens)
	}
}

func TestMultilineComment(t *testing.T) {
	state := &State{}
	line1 := Tokenize("/* start of comment", "go", state)
	if !state.InBlockComment {
		t.Errorf("expected InBlockComment to be true after line1")
	}
	if len(line1) == 0 || line1[0].Type != TokenComment {
		t.Errorf("expected line1 to be comment, got %+v", line1)
	}

	line2 := Tokenize("still in comment */ func end()", "go", state)
	if state.InBlockComment {
		t.Errorf("expected InBlockComment to be false after line2")
	}

	hasFunc := false
	for _, tok := range line2 {
		if tok.Type == TokenKeyword && tok.Value == "func" {
			hasFunc = true
		}
	}
	if !hasFunc {
		t.Errorf("expected func after multiline comment close, got %+v", line2)
	}
}

func TestHighlightLine(t *testing.T) {
	th := theme.ThemeDefault
	state := &State{}

	// No color profile
	plain := HighlightLine("func test()", "go", state, th, terminal.ColorNone)
	if plain != "func test()" {
		t.Errorf("expected unformatted plain text, got %q", plain)
	}

	// TrueColor profile
	colored := HighlightLine("func test()", "go", state, th, terminal.ColorTrueColor)
	if !strings.Contains(colored, "\x1b[") {
		t.Errorf("expected ANSI escape codes in TrueColor output, got %q", colored)
	}
}
