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

// ContextExpression represents a reference to a value in the context
type ContextExpression struct {
	Path []string // Path to the value in the context object
}

func (e *ContextExpression) String() string {
	return strings.Join(e.Path, ".")
}

// AttributeExpression represents a reference to an entity attribute
type AttributeExpression struct {
	EntityType   string // The entity type that has the attribute
	AttributeName string // The attribute name
}

func (e *AttributeExpression) String() string {
	if e.EntityType == "" {
		return e.AttributeName
	}
	return fmt.Sprintf("%s.%s", e.EntityType, e.AttributeName)
}

// RuleExpression represents a call to a rule function
type RuleExpression struct {
	RuleName  string       // The name of the rule
	Arguments []Expression // The arguments to the rule
}

func (e *RuleExpression) String() string {
	args := make([]string, len(e.Arguments))
	for i, arg := range e.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", e.RuleName, strings.Join(args, ", "))
}

// LiteralExpression represents a literal value (string, number, boolean)
type LiteralExpression struct {
	Value interface{}
}

func (e *LiteralExpression) String() string {
	switch v := e.Value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", e.Value)
	}
}

// ComparisonExpression represents a comparison between two expressions
type ComparisonExpression struct {
	Left     Expression
	Operator tokenType
	Right    Expression
}

func (e *ComparisonExpression) String() string {
	var op string
	switch e.Operator {
	case tokenEQ:
		op = "=="
	case tokenNEQ:
		op = "!="
	case tokenGT:
		op = ">"
	case tokenGTE:
		op = ">="
	case tokenLT:
		op = "<"
	case tokenLTE:
		op = "<="
	default:
		op = "??"
	}
	return fmt.Sprintf("%s %s %s", e.Left.String(), op, e.Right.String())
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
	tokenComma
	tokenEQ  // ==
	tokenNEQ // !=
	tokenGT  // >
	tokenGTE // >=
	tokenLT  // <
	tokenLTE // <=
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

		case input[pos] == ',':
			p.tokens = append(p.tokens, Token{Type: tokenComma, Value: ","})
			pos++

		case input[pos] == '.':
			p.tokens = append(p.tokens, Token{Type: tokenDot, Value: "."})
			pos++

		case input[pos] == '=':
			if pos+1 < len(input) && input[pos+1] == '=' {
				p.tokens = append(p.tokens, Token{Type: tokenEQ, Value: "=="})
				pos += 2
			} else {
				return fmt.Errorf("unexpected character: %c at position %d (expected '==')", input[pos], pos)
			}

		case input[pos] == '!':
			if pos+1 < len(input) && input[pos+1] == '=' {
				p.tokens = append(p.tokens, Token{Type: tokenNEQ, Value: "!="})
				pos += 2
			} else {
				return fmt.Errorf("unexpected character: %c at position %d (expected '!=')", input[pos], pos)
			}

		case input[pos] == '>':
			if pos+1 < len(input) && input[pos+1] == '=' {
				p.tokens = append(p.tokens, Token{Type: tokenGTE, Value: ">="})
				pos += 2
			} else {
				p.tokens = append(p.tokens, Token{Type: tokenGT, Value: ">"})
				pos++
			}

		case input[pos] == '<':
			if pos+1 < len(input) && input[pos+1] == '=' {
				p.tokens = append(p.tokens, Token{Type: tokenLTE, Value: "<="})
				pos += 2
			} else {
				p.tokens = append(p.tokens, Token{Type: tokenLT, Value: "<"})
				pos++
			}

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
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(tokenAnd) {
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &AndExpression{Left: left, Right: right}
	}

	return left, nil
}

// parseComparison parses comparison expressions (==, !=, >, >=, <, <=)
func (p *ConditionParser) parseComparison() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Check for comparison operators
	if p.match(tokenEQ, tokenNEQ, tokenGT, tokenGTE, tokenLT, tokenLTE) {
		operator := p.previous().Type
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		// Create a comparison expression based on the operator
		return &ComparisonExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		}, nil
	}

	return left, nil
}

// parsePrimary parses primary expressions (identifiers, literals, or parenthesized expressions)
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

	// Parse identifier (relation, attribute, rule call, or context reference)
	if p.check(tokenIdentifier) {
		identName := p.advance().Value

		// Check for function call (rule call)
		if p.match(tokenLeftParen) {
			// This is a rule call
			ruleExpr := &RuleExpression{
				RuleName:  identName,
				Arguments: []Expression{},
			}

			// Parse arguments if there are any
			if !p.check(tokenRightParen) {
				for {
					// Parse the argument expression
					argExpr, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					ruleExpr.Arguments = append(ruleExpr.Arguments, argExpr)

					// Check if there are more arguments
					if !p.match(tokenComma) {
						break
					}
				}
			}

			// Expect closing parenthesis
			if !p.match(tokenRightParen) {
				return nil, fmt.Errorf("expected closing parenthesis for rule call")
			}

			return ruleExpr, nil
		}

		// Check for dot notation (e.g., "organization.owner" or "request.amount")
		if p.match(tokenDot) {
			if !p.check(tokenIdentifier) {
				return nil, fmt.Errorf("expected identifier after dot")
			}

			secondPart := p.advance().Value

			// Special handling for request context
			if identName == "request" {
				return &ContextExpression{
					Path: []string{identName, secondPart},
				}, nil
			}

			// For now, treat all other dotted references as relation references
			// In a complete implementation, this would check the schema to determine
			// if this is a relation or attribute reference
			return &RelationExpression{
				RelationPath: identName,
				RelationName: secondPart,
			}, nil
		}

		// Simple identifier - treat as relation reference
		// In a complete implementation, this would check the schema to determine
		// if this is a relation or attribute reference
		return &RelationExpression{
			RelationPath: "",
			RelationName: identName,
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
