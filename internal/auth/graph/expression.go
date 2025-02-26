package graph

import (
	"fmt"
	"strings"
	"unicode"
)

// Expression interface for all condition expressions
type Expression interface {
	String() string
}

// AndExpression represents logical AND of two expressions
type AndExpression struct {
	Left  Expression
	Right Expression
}

func (e *AndExpression) String() string {
	return fmt.Sprintf("(%s and %s)", e.Left.String(), e.Right.String())
}

// OrExpression represents logical OR of two expressions
type OrExpression struct {
	Left  Expression
	Right Expression
}

func (e *OrExpression) String() string {
	return fmt.Sprintf("(%s or %s)", e.Left.String(), e.Right.String())
}

// RelationExpression represents a direct or indirect relation check
type RelationExpression struct {
	RelationPath string // Empty for direct relations, entity type for indirect
	RelationName string // The actual relation name
}

func (e *RelationExpression) String() string {
	if e.RelationPath == "" {
		return e.RelationName
	}
	return fmt.Sprintf("%s.%s", e.RelationPath, e.RelationName)
}

// Token types for the lexer
type tokenType int

const (
	tokenIdentifier tokenType = iota
	tokenDot
	tokenAnd
	tokenOr
	tokenLeftParen
	tokenRightParen
	tokenEOF
)

// Token represents a lexical token
type Token struct {
	Type  tokenType
	Value string
}

// ConditionParser parses boolean expressions for permission conditions
type ConditionParser struct {
	input   string
	tokens  []Token
	current int
}

// NewConditionParser creates a new parser for condition expressions
func NewConditionParser(input string) *ConditionParser {
	return &ConditionParser{
		input:   input,
		tokens:  []Token{},
		current: 0,
	}
}

// Parse tokenizes and parses the condition expression
func (p *ConditionParser) Parse() (Expression, error) {
	// Tokenize the input
	if err := p.tokenize(); err != nil {
		return nil, err
	}

	// Parse the tokens into an expression tree
	return p.parseExpression()
}

// tokenize breaks the input string into tokens
func (p *ConditionParser) tokenize() error {
	input := p.input
	pos := 0

	for pos < len(input) {
		switch {
		case unicode.IsSpace(rune(input[pos])):
			// Skip whitespace
			pos++

		case input[pos] == '(':
			p.tokens = append(p.tokens, Token{Type: tokenLeftParen, Value: "("})
			pos++

		case input[pos] == ')':
			p.tokens = append(p.tokens, Token{Type: tokenRightParen, Value: ")"})
			pos++

		case input[pos] == '.':
			p.tokens = append(p.tokens, Token{Type: tokenDot, Value: "."})
			pos++

		case unicode.IsLetter(rune(input[pos])) || unicode.IsDigit(rune(input[pos])) || input[pos] == '_':
			// Parse identifier (relation name or operator)
			start := pos
			for pos < len(input) && (unicode.IsLetter(rune(input[pos])) ||
				unicode.IsDigit(rune(input[pos])) || input[pos] == '_') {
				pos++
			}
			word := input[start:pos]

			// Check if it's a keyword or an identifier
			switch strings.ToLower(word) {
			case "and":
				p.tokens = append(p.tokens, Token{Type: tokenAnd, Value: word})
			case "or":
				p.tokens = append(p.tokens, Token{Type: tokenOr, Value: word})
			default:
				p.tokens = append(p.tokens, Token{Type: tokenIdentifier, Value: word})
			}

		default:
			return fmt.Errorf("unexpected character: %c at position %d", input[pos], pos)
		}
	}

	// Add EOF token
	p.tokens = append(p.tokens, Token{Type: tokenEOF})
	return nil
}

// parseExpression parses a boolean expression
func (p *ConditionParser) parseExpression() (Expression, error) {
	// Start with OR-level precedence
	return p.parseOr()
}

// parseOr parses expressions connected with OR
func (p *ConditionParser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(tokenOr) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &OrExpression{Left: left, Right: right}
	}

	return left, nil
}

// parseAnd parses expressions connected with AND
func (p *ConditionParser) parseAnd() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.match(tokenAnd) {
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &AndExpression{Left: left, Right: right}
	}

	return left, nil
}

// parsePrimary parses primary expressions (identifiers or parenthesized expressions)
func (p *ConditionParser) parsePrimary() (Expression, error) {
	// Parse parenthesized expression
	if p.match(tokenLeftParen) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.match(tokenRightParen) {
			return nil, fmt.Errorf("expected closing parenthesis")
		}

		return expr, nil
	}

	// Parse relation expression
	if p.check(tokenIdentifier) {
		entityType := ""
		relationName := p.advance().Value

		// Check for dot notation (e.g., "organization.owner")
		if p.match(tokenDot) {
			if !p.check(tokenIdentifier) {
				return nil, fmt.Errorf("expected relation name after dot")
			}

			entityType = relationName
			relationName = p.advance().Value
		}

		return &RelationExpression{
			RelationPath: entityType,
			RelationName: relationName,
		}, nil
	}

	return nil, fmt.Errorf("unexpected token: %v", p.peek())
}

// Helper methods for the parser

// match consumes the current token if it matches the given type
func (p *ConditionParser) match(types ...tokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// check tests if the current token is of the given type
func (p *ConditionParser) check(tokenType tokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

// advance consumes the current token and returns it
func (p *ConditionParser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we've consumed all tokens
func (p *ConditionParser) isAtEnd() bool {
	return p.peek().Type == tokenEOF
}

// peek returns the current token without consuming it
func (p *ConditionParser) peek() Token {
	return p.tokens[p.current]
}

// previous returns the most recently consumed token
func (p *ConditionParser) previous() Token {
	return p.tokens[p.current-1]
}
