// File: parser/token.go
package parser

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// TokenType represents the type of a token
type TokenType int

// Token types
const (
	TokenIllegal TokenType = iota
	TokenEOF
	TokenComment

	// Identifiers and literals
	TokenIdent

	// Keywords
	TokenEntity
	TokenRelation
	TokenPermission
	TokenAttribute

	// Operators and delimiters
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenAt        // @
	TokenEquals    // =
	TokenOr        // or
	TokenAnd       // and
	TokenLParen    // (
	TokenRParen    // )
	TokenDot       // .
	TokenSemicolon // ;
)

// Keywords maps keyword strings to token types
var Keywords = map[string]TokenType{
	"entity":     TokenEntity,
	"relation":   TokenRelation,
	"permission": TokenPermission,
	"attribute":  TokenAttribute,
	"or":         TokenOr,
	"and":        TokenAnd,
}
