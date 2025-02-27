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
	TokenRule

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
	TokenComma     // ,
	TokenSemicolon // ;
	
	// Comparison operators
	TokenGT        // >
	TokenGTE       // >=
	TokenLT        // <
	TokenLTE       // <=
	TokenEQ        // ==
	TokenNEQ       // !=
)

// Keywords maps keyword strings to token types
var Keywords = map[string]TokenType{
	"entity":     TokenEntity,
	"relation":   TokenRelation,
	"permission": TokenPermission,
	"attribute":  TokenAttribute,
	"rule":       TokenRule,
	"or":         TokenOr,
	"and":        TokenAnd,
}
