// File: parser/lexer.go
package parser

import (
	"unicode"
)

// Lexer tokenizes input text
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	line         int
	column       int
}

// NewLexer creates a new Lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = rune(l.input[l.readPosition])
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar returns the next character without advancing the position
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	}
	return rune(l.input[l.readPosition])
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	// Skip whitespace and comments before processing the next token
	for unicode.IsSpace(l.ch) || (l.ch == '/' && l.peekChar() == '/') {
		if l.ch == '/' && l.peekChar() == '/' {
			// Skip the entire comment
			l.skipComment()
		} else {
			// Skip whitespace
			l.readChar()
		}
	}

	switch l.ch {
	case '{':
		tok = Token{Type: TokenLBrace, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '}':
		tok = Token{Type: TokenRBrace, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '@':
		tok = Token{Type: TokenAt, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '=':
		tok = Token{Type: TokenEquals, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '(':
		tok = Token{Type: TokenLParen, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: TokenRParen, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '.':
		tok = Token{Type: TokenDot, Literal: string(l.ch), Line: l.line, Column: l.column}
	case 0:
		tok = Token{Type: TokenEOF, Literal: "", Line: l.line, Column: l.column}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column - len(tok.Literal)
			return tok
		} else {
			tok = Token{Type: TokenIllegal, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	}

	l.readChar()
	return tok
}

// readIdentifier reads an identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipComment skips over a comment line
func (l *Lexer) skipComment() {
	// Skip the initial //
	l.readChar()
	l.readChar()

	// Skip until end of line or EOF
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	// Skip the newline if present
	if l.ch == '\n' {
		l.readChar()
	}
}

// isLetter returns true if the character is a letter
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// isDigit returns true if the character is a digit
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// lookupIdent checks if the identifier is a keyword
func lookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
