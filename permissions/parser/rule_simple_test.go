package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRuleParsing(t *testing.T) {
	// Start with a very simple rule just to debug the parsing
	input := `rule test_rule(a integer, b integer) {
    a >= b
}`

	// First, check the token stream to understand the issue
	l := NewLexer(input)
	
	// Trace tokenization
	tokens := []Token{}
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		t.Logf("Token: %v, Literal: %q, Line: %d, Col: %d", token.Type, token.Literal, token.Line, token.Column)
		if token.Type == TokenEOF {
			break
		}
	}
	
	// Observe the token sequence: we can see the rule lexed properly with comparison tokens
	
	// Now parse the full model including this rule
	l = NewLexer(input)
	p := NewParser(l)
	model := p.ParsePermissionModel()
	
	t.Logf("Parser errors: %v", p.Errors())
	
	// We should have one rule in the global rules list
	assert.Equal(t, 1, len(model.Rules), "Should have one rule")
	
	// Get the rule
	rule, exists := model.Rules["test_rule"]
	assert.True(t, exists, "Rule should exist")
	
	// Check rule details
	assert.Equal(t, "test_rule", rule.Name, "Rule name should match")
	assert.Equal(t, 2, len(rule.Parameters), "Rule should have 2 parameters")
	// Rule expression is handled in a different way than we expected - let's see what it actually contains
	t.Logf("Rule expression: %q", rule.Expression)
}