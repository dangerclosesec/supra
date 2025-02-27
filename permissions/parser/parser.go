// File: parser/parser.go
package parser

import (
	"fmt"

	"github.com/dangerclosesec/supra/permissions/model"
)

// Parser parses .perm files into a permission model
type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
	// We'll use a different approach for comments that doesn't interfere with parsing
	currentComments []string
}

// NewParser creates a new Parser
func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:               l,
		errors:          []string{},
		currentComments: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// ParsePermissionModel parses a complete permission model
func (p *Parser) ParsePermissionModel() *model.PermissionModel {
	permModel := model.NewPermissionModel()

	for p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenEntity {
			entity := p.parseEntity()
			if entity != nil {
				permModel.AddEntity(entity)
			}
		} else {
			// Skip any unexpected tokens at the top level
			p.nextToken()
		}
	}

	return permModel
}

// parseEntity parses an entity declaration
func (p *Parser) parseEntity() *model.Entity {
	// "entity" keyword is already consumed
	if !p.expectPeek(TokenIdent) {
		return nil
	}

	entity := &model.Entity{
		Name: p.curToken.Literal,
	}

	if !p.expectPeek(TokenLBrace) {
		return nil
	}

	p.nextToken() // Move past the opening brace

	// Parse relations, permissions, and attributes
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenRelation {
			relation := p.parseRelation()
			if relation != nil {
				entity.Relations = append(entity.Relations, *relation)
			}
		} else if p.curToken.Type == TokenPermission {
			permission := p.parsePermission()
			if permission != nil {
				entity.Permissions = append(entity.Permissions, *permission)
			}
		} else if p.curToken.Type == TokenAttribute {
			attribute := p.parseAttribute()
			if attribute != nil {
				entity.Attributes = append(entity.Attributes, *attribute)
			}
		} else {
			// Skip unexpected tokens within entity body
			p.nextToken()
		}
	}

	// Consume the closing brace
	if p.curToken.Type == TokenRBrace {
		p.nextToken()
	}

	return entity
}

// parseRelation parses a relation declaration
func (p *Parser) parseRelation() *model.Relation {
	startLine := p.curToken.Line

	// "relation" keyword is already consumed
	if !p.expectPeek(TokenIdent) {
		return nil
	}

	relation := &model.Relation{
		Name:       p.curToken.Literal,
		LineNumber: startLine,
	}

	if !p.expectPeek(TokenAt) {
		return nil
	}

	if !p.expectPeek(TokenIdent) {
		return nil
	}

	relation.Target = p.curToken.Literal

	// Consume any tokens until we reach a new statement or the end of the entity
	for p.peekToken.Type != TokenRelation &&
		p.peekToken.Type != TokenPermission &&
		p.peekToken.Type != TokenAttribute &&
		p.peekToken.Type != TokenRBrace &&
		p.peekToken.Type != TokenEOF {
		p.nextToken()
	}

	// Move to the next token to prepare for the next statement
	p.nextToken()

	return relation
}

// parsePermission parses a permission declaration
func (p *Parser) parsePermission() *model.Permission {
	startLine := p.curToken.Line

	// "permission" keyword is already consumed
	if !p.expectPeek(TokenIdent) {
		return nil
	}

	permission := &model.Permission{
		Name:       p.curToken.Literal,
		LineNumber: startLine,
	}

	if !p.expectPeek(TokenEquals) {
		return nil
	}

	p.nextToken() // Move past the equals sign

	// Parse the expression
	expr, exprStr := p.parseExpression(LOWEST)
	if expr != nil {
		permission.ParsedExpr = expr
		permission.Expression = exprStr

		// Consume any tokens until we reach a new statement or the end of the entity
		for p.peekToken.Type != TokenRelation &&
			p.peekToken.Type != TokenPermission &&
			p.peekToken.Type != TokenAttribute &&
			p.peekToken.Type != TokenRBrace &&
			p.peekToken.Type != TokenEOF {
			p.nextToken()
		}

		// Move to the next token to prepare for the next statement
		p.nextToken()

		return permission
	}

	// If we couldn't parse an expression, skip to the next statement
	for p.curToken.Type != TokenRelation &&
		p.curToken.Type != TokenPermission &&
		p.curToken.Type != TokenAttribute &&
		p.curToken.Type != TokenRBrace &&
		p.curToken.Type != TokenEOF {
		p.nextToken()
	}

	return nil
}

// parseAttribute parses an attribute declaration
func (p *Parser) parseAttribute() *model.Attribute {
	startLine := p.curToken.Line

	// "attribute" keyword is already consumed
	if !p.expectPeek(TokenIdent) {
		return nil
	}

	attributeName := p.curToken.Literal

	if !p.expectPeek(TokenIdent) {
		return nil
	}

	// The token after attribute name should be the data type
	dataTypeStr := p.curToken.Literal
	
	// Check if the next token is [ for array types
	if p.peekTokenIs(TokenLBracket) {
		p.nextToken() // consume [
		
		// Expect ]
		if !p.expectPeek(TokenRBracket) {
			p.addError("expected ']' after '['")
			return nil
		}
		
		// Modify dataTypeStr to include the array suffix
		dataTypeStr += "[]"
	}
	
	if !model.IsValidAttributeDataType(dataTypeStr) {
		p.addError(fmt.Sprintf("invalid attribute data type: %s", dataTypeStr))
		return nil
	}

	attribute := &model.Attribute{
		Name:       attributeName,
		DataType:   model.AttributeDataType(dataTypeStr),
		LineNumber: startLine,
	}

	// Consume any tokens until we reach a new statement or the end of the entity
	for p.peekToken.Type != TokenRelation &&
		p.peekToken.Type != TokenPermission &&
		p.peekToken.Type != TokenAttribute &&
		p.peekToken.Type != TokenRBrace &&
		p.peekToken.Type != TokenEOF {
		p.nextToken()
	}

	// Move to the next token to prepare for the next statement
	p.nextToken()

	return attribute
}

// parseExpression parses a permission expression
func (p *Parser) parseExpression(precedence int) (model.Expression, string) {
	var leftExpr model.Expression
	leftStr := ""

	// Parse the prefix expression
	switch p.curToken.Type {
	case TokenIdent:
		// Check if it's a dotted reference
		if p.peekTokenIs(TokenDot) {
			entity := p.curToken.Literal
			p.nextToken() // consume the dot
			p.nextToken() // move to the relation name
			if p.curToken.Type != TokenIdent {
				p.addError(fmt.Sprintf("expected identifier after dot, got %s", p.curToken.Literal))
				return nil, ""
			}
			leftExpr = &model.RelationRef{
				Entity: entity,
				Name:   p.curToken.Literal,
			}
			leftStr = entity + "." + p.curToken.Literal
		} else {
			// Direct relation reference
			leftExpr = &model.RelationRef{
				Name: p.curToken.Literal,
			}
			leftStr = p.curToken.Literal
		}
	case TokenLParen:
		p.nextToken() // Move past the opening parenthesis
		innerExpr, innerStr := p.parseExpression(LOWEST)
		if !p.expectPeek(TokenRParen) {
			p.addError("expected closing parenthesis")
			return nil, ""
		}
		leftExpr = &model.Parentheses{Expr: innerExpr}
		leftStr = "(" + innerStr + ")"
	default:
		p.addError(fmt.Sprintf("unexpected token in expression: %s", p.curToken.Literal))
		return nil, ""
	}

	// Parse infix expressions (and, or)
	for !p.peekTokenIs(TokenEOF) && precedence < p.peekPrecedence() {
		if !p.peekTokenIs(TokenAnd) && !p.peekTokenIs(TokenOr) {
			break
		}

		p.nextToken() // Move to the operator
		infix := p.curToken.Literal

		precedence := p.curPrecedence()
		p.nextToken() // Move past the operator

		rightExpr, rightStr := p.parseExpression(precedence)
		if rightExpr == nil {
			return nil, ""
		}

		if p.curTokenIs(TokenAnd) {
			leftExpr = &model.And{Left: leftExpr, Right: rightExpr}
		} else {
			leftExpr = &model.Or{Left: leftExpr, Right: rightExpr}
		}

		leftStr = leftStr + " " + infix + " " + rightStr
	}

	return leftExpr, leftStr
}

// Precedence levels
const (
	LOWEST  = 1
	OR      = 2
	AND     = 3
	PRIMARY = 4
)

// peekPrecedence returns the precedence of the peek token
func (p *Parser) peekPrecedence() int {
	switch p.peekToken.Type {
	case TokenOr:
		return OR
	case TokenAnd:
		return AND
	default:
		return LOWEST
	}
}

// curPrecedence returns the precedence of the current token
func (p *Parser) curPrecedence() int {
	switch p.curToken.Type {
	case TokenOr:
		return OR
	case TokenAnd:
		return AND
	default:
		return LOWEST
	}
}

// Helper methods for token checking
func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the expected type and advances if so
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected next token to be %v, got %v instead", t, p.peekToken.Type))
	return false
}

// addError adds an error to the parser errors
func (p *Parser) addError(msg string) {
	errorMsg := fmt.Sprintf("%s (line %d, column %d)", msg, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, errorMsg)
}
