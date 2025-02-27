package parser

import (
	"testing"

	"github.com/dangerclosesec/supra/permissions/model"
	"github.com/stretchr/testify/assert"
)

func TestRuleParsing(t *testing.T) {
	input := `rule check_balance(balance double, amount double) {
    balance >= amount
}

rule check_limit(withdraw_limit double, amount double) {
    withdraw_limit >= amount 
}

rule check_admin_approval(approval_num integer, admin_approval_limit integer) {
    approval_num >= admin_approval_limit
}

entity account {
    relation owner @organization
    attribute balance double
    attribute withdraw_limit double

    permission withdraw = check_balance(balance, request.amount) and (check_limit(withdraw_limit, request.amount) or owner.approval)
}`

	l := NewLexer(input)
	p := NewParser(l)
	permModel := p.ParsePermissionModel()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Check that rules were parsed
	assert.Equal(t, 3, len(permModel.Rules), "Should have parsed 3 rules")

	// Check that check_balance rule was parsed correctly
	rule, exists := permModel.Rules["check_balance"]
	assert.True(t, exists, "check_balance rule should exist")
	assert.Equal(t, "check_balance", rule.Name, "Rule name should match")
	assert.Equal(t, 2, len(rule.Parameters), "Rule should have 2 parameters")
	assert.Equal(t, "balance", rule.Parameters[0].Name, "First parameter name should be balance")
	assert.Equal(t, model.AttributeTypeDouble, rule.Parameters[0].DataType, "First parameter type should be double")
	assert.Equal(t, "amount", rule.Parameters[1].Name, "Second parameter name should be amount")
	assert.Equal(t, model.AttributeTypeDouble, rule.Parameters[1].DataType, "Second parameter type should be double")
	assert.Equal(t, "balance >= amount", rule.Expression, "Rule expression should match")

	// Check that check_limit rule was parsed correctly
	rule, exists = permModel.Rules["check_limit"]
	assert.True(t, exists, "check_limit rule should exist")
	assert.Equal(t, "check_limit", rule.Name, "Rule name should match")
	assert.Equal(t, 2, len(rule.Parameters), "Rule should have 2 parameters")
	assert.Equal(t, "withdraw_limit", rule.Parameters[0].Name, "First parameter name should be withdraw_limit")
	assert.Equal(t, model.AttributeTypeDouble, rule.Parameters[0].DataType, "First parameter type should be double")
	assert.Equal(t, "amount", rule.Parameters[1].Name, "Second parameter name should be amount")
	assert.Equal(t, model.AttributeTypeDouble, rule.Parameters[1].DataType, "Second parameter type should be double")
	assert.Equal(t, "withdraw_limit >= amount", rule.Expression, "Rule expression should match")

	// Check that entity's permission uses rule calls
	entity, exists := permModel.Entities["account"]
	assert.True(t, exists, "account entity should exist")
	assert.Equal(t, 1, len(entity.Permissions), "Entity should have 1 permission")
	
	permission := entity.Permissions[0]
	assert.Equal(t, "withdraw", permission.Name, "Permission name should match")
	assert.Equal(t, "check_balance(balance, request.amount) and (check_limit(withdraw_limit, request.amount) or owner.approval)", permission.Expression, 
		"Permission expression should include rule calls")
}

func TestRuleInEntityParsing(t *testing.T) {
	input := `entity account {
    relation owner @organization
    attribute balance double
    attribute withdraw_limit double

    rule check_balance(balance double, amount double) {
        balance >= amount
    }
    
    rule check_limit(withdraw_limit double, amount double) {
        withdraw_limit >= amount 
    }

    permission withdraw = check_balance(balance, request.amount) and (check_limit(withdraw_limit, request.amount) or owner.approval)
}`

	l := NewLexer(input)
	p := NewParser(l)
	permModel := p.ParsePermissionModel()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Check that the entity was parsed with rules
	entity, exists := permModel.Entities["account"]
	assert.True(t, exists, "account entity should exist")
	assert.Equal(t, 2, len(entity.Rules), "Entity should have 2 rules")
	
	// Check that entity rules were also added to global rules
	assert.Equal(t, 2, len(permModel.Rules), "PermissionModel should have 2 global rules")

	// Check that check_balance rule was parsed correctly
	rule := entity.Rules[0]
	assert.Equal(t, "check_balance", rule.Name, "Rule name should match")
	assert.Equal(t, 2, len(rule.Parameters), "Rule should have 2 parameters")
	assert.Equal(t, "balance >= amount", rule.Expression, "Rule expression should match")

	// Check that check_limit rule was parsed correctly
	rule = entity.Rules[1]
	assert.Equal(t, "check_limit", rule.Name, "Rule name should match")
	assert.Equal(t, 2, len(rule.Parameters), "Rule should have 2 parameters")
	assert.Equal(t, "withdraw_limit >= amount", rule.Expression, "Rule expression should match")
}

func TestComparisonOperatorsInRules(t *testing.T) {
	input := `rule equals_check(a integer, b integer) {
    a == b
}

rule not_equals_check(a integer, b integer) {
    a != b
}

rule greater_than_check(a integer, b integer) {
    a > b
}

rule greater_than_equals_check(a integer, b integer) {
    a >= b
}

rule less_than_check(a integer, b integer) {
    a < b
}

rule less_than_equals_check(a integer, b integer) {
    a <= b
}`

	l := NewLexer(input)
	p := NewParser(l)
	permModel := p.ParsePermissionModel()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Check that all rules were parsed
	assert.Equal(t, 6, len(permModel.Rules), "Should have parsed all 6 rules")

	// Check each rule's expression
	comparisons := map[string]string{
		"equals_check":             "a == b",
		"not_equals_check":         "a != b",
		"greater_than_check":       "a > b",
		"greater_than_equals_check": "a >= b",
		"less_than_check":          "a < b",
		"less_than_equals_check":   "a <= b",
	}

	for ruleName, expectedExpr := range comparisons {
		rule, exists := permModel.Rules[ruleName]
		assert.True(t, exists, "%s rule should exist", ruleName)
		assert.Equal(t, expectedExpr, rule.Expression, "%s rule expression should match", ruleName)
	}
}