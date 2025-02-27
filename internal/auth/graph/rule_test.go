package graph

import (
	"fmt"
	"strings"
	"testing"
)

func TestRuleEvaluationBasics(t *testing.T) {
	// We'll create a simpler test that just verifies the context path extraction logic
	
	// Sample context data for testing
	contextData := map[string]interface{}{
		"request": map[string]interface{}{
			"authorized": true,
			"user": map[string]interface{}{
				"admin": true,
				"roles": []string{"editor", "viewer"},
			},
			"amount": 50.0,
			"limit":  100.0,
		},
	}
	
	// Test navigating context paths
	tests := []struct {
		path  []string
		exists bool
	}{
		{[]string{"request", "authorized"}, true},
		{[]string{"request", "user", "admin"}, true},
		{[]string{"request", "missing"}, false},
		{[]string{"request", "amount"}, true},
		{[]string{"nonexistent"}, false},
	}
	
	for _, tc := range tests {
		t.Run(strings.Join(tc.path, "."), func(t *testing.T) {
			// Navigate the path
			var value interface{} = contextData
			var exists bool = true
			
			for i, part := range tc.path {
				// Try to extract the value from the current map
				mapValue, ok := value.(map[string]interface{})
				if !ok {
					exists = false
					break
				}
				
				value, exists = mapValue[part]
				if !exists {
					break
				}
				
				// If this isn't the last part and the value isn't a map, we can't continue
				if i < len(tc.path)-1 && value != nil {
					_, ok = value.(map[string]interface{})
					if !ok {
						exists = false
						break
					}
				}
			}
			
			// Check if the result matches expectation
			if exists != tc.exists {
				t.Errorf("Path %v: expected exists=%v, got %v", tc.path, tc.exists, exists)
			}
		})
	}
	
	// Test conversion to numeric types
	numericTests := []struct {
		value    interface{}
		expected float64
		valid    bool
	}{
		{42, 42.0, true},
		{42.5, 42.5, true},
		{int64(100), 100.0, true},
		{"123.45", 123.45, true},
		{"not a number", 0.0, false},
		{true, 0.0, false},
		{nil, 0.0, false},
	}
	
	for i, tc := range numericTests {
		t.Run(fmt.Sprintf("numeric conversion %d", i), func(t *testing.T) {
			result, ok := toFloat64(tc.value)
			
			if ok != tc.valid {
				t.Errorf("Expected valid=%v, got %v for value %v", tc.valid, ok, tc.value)
			}
			
			if ok && result != tc.expected {
				t.Errorf("Expected %v, got %v for value %v", tc.expected, result, tc.value)
			}
		})
	}
}

// Helper function to get value from context based on expression
func getContextValue(expr Expression, contextData map[string]interface{}) interface{} {
	if contextExpr, ok := expr.(*ContextExpression); ok {
		current := contextData
		for i, part := range contextExpr.Path {
			if i == len(contextExpr.Path)-1 {
				return current[part]
			}
			
			next, exists := current[part]
			if !exists {
				return nil
			}
			
			nextMap, ok := next.(map[string]interface{})
			if !ok {
				return nil
			}
			
			current = nextMap
		}
	}
	return nil
}