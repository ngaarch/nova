package syntax

// TokenType represents a semantic syntactic category.
type TokenType int

const (
	TokenText TokenType = iota
	TokenKeyword
	TokenTypeIdent
	TokenString
	TokenNumber
	TokenComment
	TokenOperator
	TokenPunctuation
	TokenTag
	TokenAttr
	TokenHeading
	TokenListMarker
	TokenBlockQuote
)

// Token represents a single lexical token.
type Token struct {
	Type  TokenType
	Value string
}

// State maintains cross-line parsing state for multi-line constructs.
type State struct {
	InBlockComment   bool
	BlockCommentEnd  string
	InRawString      bool
	RawStringDelim   string
	InCodeFence      bool
	CodeFenceLang    string
}
