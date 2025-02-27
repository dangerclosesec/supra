// File: parser/parser.go
package parser

import (
	"fmt"
	"strings"

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
		} else if p.curToken.Type == TokenRule {
			rule := p.parseRule()
			if rule != nil {
				// Add the rule to the global rules map
				permModel.AddRule(rule)
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

	// Parse relations, permissions, attributes, and rules
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
		} else if p.curToken.Type == TokenRule {
			rule := p.parseRule()
			if rule != nil {
				entity.Rules = append(entity.Rules, *rule)
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
		p.peekToken.Type != TokenRule &&
		p.peekToken.Type != TokenRBrace &&
		p.peekToken.Type != TokenEOF {
		p.nextToken()
	}

	// Move to the next token to prepare for the next statement
	p.nextToken()

	return attribute
}

// parseRule parses a rule declaration
func (p *Parser) parseRule() *model.Rule {
	startLine := p.curToken.Line

	// "rule" keyword is already consumed
	if !p.expectPeek(TokenIdent) {
		return nil
	}

	rule := &model.Rule{
		Name:       p.curToken.Literal,
		LineNumber: startLine,
	}

	// Parse rule parameters in parentheses
	if !p.expectPeek(TokenLParen) {
		return nil
	}
	
	// Parse parameters until we reach closing parenthesis
	for !p.peekTokenIs(TokenRParen) && !p.peekTokenIs(TokenEOF) {
		p.nextToken() // Move to parameter name
		
		if p.curToken.Type != TokenIdent {
			p.addError("expected parameter name")
			return nil
		}
		
		paramName := p.curToken.Literal
		
		if !p.expectPeek(TokenIdent) {
			p.addError("expected parameter type")
			return nil
		}
		
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
			p.addError(fmt.Sprintf("invalid parameter data type: %s", dataTypeStr))
			return nil
		}
		
		parameter := model.RuleParameter{
			Name:     paramName,
			DataType: model.AttributeDataType(dataTypeStr),
		}
		
		rule.Parameters = append(rule.Parameters, parameter)
		
		// If next token is comma, consume it and continue
		if p.peekTokenIs(TokenComma) {
			p.nextToken()
		} else if !p.peekTokenIs(TokenRParen) {
			p.addError("expected ',' or ')' after parameter")
			return nil
		}
	}
	
	// Consume the closing parenthesis
	if !p.expectPeek(TokenRParen) {
		return nil
	}
	
	// Expect opening brace
	if !p.expectPeek(TokenLBrace) {
		return nil
	}
	
	// Parse rule body (expression)
	p.nextToken() // Move past the opening brace
	
	// Instead of trying to parse the expression as a full expression tree,
	// just capture the rule body as a string for now
	var bodyBuilder strings.Builder
	
	// Read tokens until we find the closing brace
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		// Add the token to the body
		if len(p.curToken.Literal) > 0 {
			if bodyBuilder.Len() > 0 {
				bodyBuilder.WriteString(" ")
			}
			bodyBuilder.WriteString(p.curToken.Literal)
		}
		p.nextToken()
	}
	
	// Set the expression as the raw body text
	rule.Expression = strings.TrimSpace(bodyBuilder.String())
	
	// For now, use a RelationRef as a placeholder for the parsed expression
	rule.ParsedExpr = &model.RelationRef{
		Name: rule.Expression,
	}
	
	// Check if we found the closing brace
	if p.curToken.Type != TokenRBrace {
		p.addError("expected closing brace for rule body")
		return nil
	}
	
	// Move to the next token
	p.nextToken()
	
	return rule
}

// parseRuleExpression parses the expression within a rule
func (p *Parser) parseRuleExpression() (model.Expression, string) {
	// For simplicity in testing, just capture the expression as a string
	// and create a basic expression structure
	
	var expr model.Expression
	var exprStr string
	
	// For comparison expressions like "a >= b", build a relation expression
	if p.curTokenIs(TokenIdent) {
		leftName := p.curToken.Literal
		
		// Check for comparison operators
		if p.peekTokenIs(TokenGT) || p.peekTokenIs(TokenGTE) || p.peekTokenIs(TokenLT) || 
		   p.peekTokenIs(TokenLTE) || p.peekTokenIs(TokenEQ) || p.peekTokenIs(TokenNEQ) {
			
			// Get the operator
			p.nextToken()
			operator := p.curToken.Literal
			
			// Get the right side
			if p.expectPeek(TokenIdent) {
				rightName := p.curToken.Literal
				
				// Create a relation expression with the full expression as the name
				exprStr = leftName + " " + operator + " " + rightName
				expr = &model.RelationRef{
					Name: exprStr,
				}
				
				// Move to the next token
				p.nextToken()
				return expr, exprStr
			}
		}
	}
	
	// If not a comparison expression, fall back to regular expression parsing
	expr, exprStr = p.parseExpression(LOWEST)
	return expr, exprStr
}

// parseExpression parses a permission expression
func (p *Parser) parseExpression(precedence int) (model.Expression, string) {
	var leftExpr model.Expression
	leftStr := ""

	// Parse the prefix expression
	switch p.curToken.Type {
	case TokenIdent:
		// Check if it's a function call (rule call)
		if p.peekTokenIs(TokenLParen) {
			ruleName := p.curToken.Literal
			p.nextToken() // consume the left paren
			
			// Parse rule arguments
			ruleCall := &model.RuleCall{
				Name: ruleName,
			}
			
			argStrings := []string{}
			
			// If there are arguments
			if !p.peekTokenIs(TokenRParen) {
				p.nextToken() // Move to the first argument
				
				for {
					argExpr, argStr := p.parseExpression(LOWEST)
					if argExpr == nil {
						p.addError("invalid rule argument")
						return nil, ""
					}
					
					ruleCall.Arguments = append(ruleCall.Arguments, argExpr)
					argStrings = append(argStrings, argStr)
					
					// If next token is comma, continue to next argument
					if p.peekTokenIs(TokenComma) {
						p.nextToken() // consume comma
						p.nextToken() // move to next argument
					} else {
						break
					}
				}
			}
			
			if !p.expectPeek(TokenRParen) {
				p.addError("expected closing parenthesis for rule call")
				return nil, ""
			}
			
			leftExpr = ruleCall
			leftStr = ruleName + "(" + strings.Join(argStrings, ", ") + ")"
		} else if p.peekTokenIs(TokenDot) {
			// Check if it's a dotted reference (could be relation or attribute or request context)
			entity := p.curToken.Literal
			p.nextToken() // consume the dot
			p.nextToken() // move to the relation/attribute name
			
			if p.curToken.Type != TokenIdent {
				p.addError(fmt.Sprintf("expected identifier after dot, got %s", p.curToken.Literal))
				return nil, ""
			}
			
			// Special handling for 'request' context reference
			if entity == "request" {
				contextRef := &model.ContextRef{
					Path: []string{"request", p.curToken.Literal},
				}
				leftExpr = contextRef
				leftStr = entity + "." + p.curToken.Literal
			} else {
				// For now, treat as relation reference. In a complete implementation,
				// we would need to check if it's an attribute reference based on the model
				leftExpr = &model.RelationRef{
					Entity: entity,
					Name:   p.curToken.Literal,
				}
				leftStr = entity + "." + p.curToken.Literal
			}
		} else {
			// Try to determine if this is a relation reference or an attribute reference
			// For now, default to relation reference
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
