package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Entity represents a permify entity type
type Entity struct {
	Name        string
	Relations   []Relation
	Permissions []Permission
	Attributes  []Attribute
	Rules       []Rule
	Comments    []string
}

// AttributeDataType defines the supported data types for attributes
type AttributeDataType string

// Supported data types for attributes
const (
	AttributeTypeBoolean      AttributeDataType = "boolean"
	AttributeTypeBooleanArray AttributeDataType = "boolean[]"
	AttributeTypeString       AttributeDataType = "string"
	AttributeTypeStringArray  AttributeDataType = "string[]"
	AttributeTypeInteger      AttributeDataType = "integer"
	AttributeTypeIntegerArray AttributeDataType = "integer[]"
	AttributeTypeDouble       AttributeDataType = "double"
	AttributeTypeDoubleArray  AttributeDataType = "double[]"
)

// Attribute represents an entity attribute definition
type Attribute struct {
	Name       string           // Name of the attribute
	DataType   AttributeDataType // Data type of the attribute
	Comments   []string         // Documentation comments
	LineNumber int              // Line number in the source file
}

// AttributeValue represents the value of an attribute for a specific entity
type AttributeValue struct {
	Value interface{} // Actual value of the attribute
}

// ParseAttributeValue parses a string value into the appropriate type based on the attribute data type
func ParseAttributeValue(dataType AttributeDataType, value string) (interface{}, error) {
	switch dataType {
	case AttributeTypeBoolean:
		return strings.ToLower(value) == "true", nil
	case AttributeTypeBooleanArray:
		var values []bool
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return nil, fmt.Errorf("invalid boolean array: %w", err)
		}
		return values, nil
	case AttributeTypeString:
		return value, nil
	case AttributeTypeStringArray:
		var values []string
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return nil, fmt.Errorf("invalid string array: %w", err)
		}
		return values, nil
	case AttributeTypeInteger:
		var intValue int64
		if err := json.Unmarshal([]byte(value), &intValue); err != nil {
			return nil, fmt.Errorf("invalid integer: %w", err)
		}
		return intValue, nil
	case AttributeTypeIntegerArray:
		var values []int64
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return nil, fmt.Errorf("invalid integer array: %w", err)
		}
		return values, nil
	case AttributeTypeDouble:
		var doubleValue float64
		if err := json.Unmarshal([]byte(value), &doubleValue); err != nil {
			return nil, fmt.Errorf("invalid double: %w", err)
		}
		return doubleValue, nil
	case AttributeTypeDoubleArray:
		var values []float64
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return nil, fmt.Errorf("invalid double array: %w", err)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported attribute data type: %s", dataType)
	}
}

// ValidateAttributeValue validates that a value is of the appropriate type for the attribute
func ValidateAttributeValue(dataType AttributeDataType, value interface{}) error {
	// Handle nil values
	if value == nil {
		return fmt.Errorf("nil value not allowed for attribute")
	}

	valueType := reflect.TypeOf(value)
	valueKind := valueType.Kind()

	switch dataType {
	case AttributeTypeBoolean:
		if valueKind != reflect.Bool {
			return fmt.Errorf("expected boolean, got %v", valueKind)
		}
	case AttributeTypeBooleanArray:
		if valueKind != reflect.Slice {
			return fmt.Errorf("expected array, got %v", valueKind)
		}
		// Check each element is a boolean
		valueSlice := reflect.ValueOf(value)
		for i := 0; i < valueSlice.Len(); i++ {
			if valueSlice.Index(i).Kind() != reflect.Bool {
				return fmt.Errorf("expected boolean array element, got %v", valueSlice.Index(i).Kind())
			}
		}
	case AttributeTypeString:
		if valueKind != reflect.String {
			return fmt.Errorf("expected string, got %v", valueKind)
		}
	case AttributeTypeStringArray:
		if valueKind != reflect.Slice {
			return fmt.Errorf("expected array, got %v", valueKind)
		}
		// Check each element is a string
		valueSlice := reflect.ValueOf(value)
		for i := 0; i < valueSlice.Len(); i++ {
			if valueSlice.Index(i).Kind() != reflect.String {
				return fmt.Errorf("expected string array element, got %v", valueSlice.Index(i).Kind())
			}
		}
	case AttributeTypeInteger:
		if valueKind != reflect.Int && valueKind != reflect.Int64 && valueKind != reflect.Int32 {
			return fmt.Errorf("expected integer, got %v", valueKind)
		}
	case AttributeTypeIntegerArray:
		if valueKind != reflect.Slice {
			return fmt.Errorf("expected array, got %v", valueKind)
		}
		// Check each element is an integer
		valueSlice := reflect.ValueOf(value)
		for i := 0; i < valueSlice.Len(); i++ {
			elemKind := valueSlice.Index(i).Kind()
			if elemKind != reflect.Int && elemKind != reflect.Int64 && elemKind != reflect.Int32 {
				return fmt.Errorf("expected integer array element, got %v", elemKind)
			}
		}
	case AttributeTypeDouble:
		if valueKind != reflect.Float64 && valueKind != reflect.Float32 {
			return fmt.Errorf("expected double, got %v", valueKind)
		}
	case AttributeTypeDoubleArray:
		if valueKind != reflect.Slice {
			return fmt.Errorf("expected array, got %v", valueKind)
		}
		// Check each element is a double
		valueSlice := reflect.ValueOf(value)
		for i := 0; i < valueSlice.Len(); i++ {
			elemKind := valueSlice.Index(i).Kind()
			if elemKind != reflect.Float64 && elemKind != reflect.Float32 {
				return fmt.Errorf("expected double array element, got %v", elemKind)
			}
		}
	default:
		return fmt.Errorf("unsupported attribute data type: %s", dataType)
	}

	return nil
}

// IsValidAttributeDataType checks if the given string is a valid attribute data type
func IsValidAttributeDataType(dataType string) bool {
	switch AttributeDataType(dataType) {
	case AttributeTypeBoolean, AttributeTypeBooleanArray,
		AttributeTypeString, AttributeTypeStringArray,
		AttributeTypeInteger, AttributeTypeIntegerArray,
		AttributeTypeDouble, AttributeTypeDoubleArray:
		return true
	default:
		return false
	}
}

// Relation represents a relationship to another entity
type Relation struct {
	Name       string
	Target     string // The entity type this relation refers to
	Comments   []string
	LineNumber int
}

// Permission represents an access control rule
type Permission struct {
	Name       string
	Expression string // The raw expression string
	ParsedExpr Expression
	Comments   []string
	LineNumber int
}

// Expression interface for permission expressions
type Expression interface {
	String() string
}

// And represents a logical AND of two expressions
type And struct {
	Left  Expression
	Right Expression
}

func (a *And) String() string {
	return "(" + a.Left.String() + " and " + a.Right.String() + ")"
}

// Or represents a logical OR of two expressions
type Or struct {
	Left  Expression
	Right Expression
}

func (o *Or) String() string {
	return "(" + o.Left.String() + " or " + o.Right.String() + ")"
}

// RelationRef refers to a relation, can be direct or nested (e.g., organization.owner)
type RelationRef struct {
	Entity string // Empty for direct relations
	Name   string
}

func (r *RelationRef) String() string {
	if r.Entity == "" {
		return r.Name
	}
	return r.Entity + "." + r.Name
}

// AttributeRef refers to an attribute, can be direct or nested (e.g., organization.premium)
type AttributeRef struct {
	Entity string // Empty for direct attributes
	Name   string
}

func (a *AttributeRef) String() string {
	if a.Entity == "" {
		return a.Name
	}
	return a.Entity + "." + a.Name
}

// Parentheses represents a parenthesized expression
type Parentheses struct {
	Expr Expression
}

func (p *Parentheses) String() string {
	return "(" + p.Expr.String() + ")"
}

// Rule represents a named function that evaluates attributes against context values
type Rule struct {
	Name       string
	Parameters []RuleParameter
	Expression string // The raw expression string
	ParsedExpr Expression
	Comments   []string
	LineNumber int
}

// RuleParameter represents a parameter passed to a rule
type RuleParameter struct {
	Name     string
	DataType AttributeDataType
}

// RuleCall represents a call to a rule with arguments
type RuleCall struct {
	Name      string
	Arguments []Expression
}

func (r *RuleCall) String() string {
	args := make([]string, len(r.Arguments))
	for i, arg := range r.Arguments {
		args[i] = arg.String()
	}
	return r.Name + "(" + strings.Join(args, ", ") + ")"
}

// ContextRef represents a reference to a value in the context
type ContextRef struct {
	Path []string
}

func (c *ContextRef) String() string {
	return strings.Join(c.Path, ".")
}

// LiteralValue represents a literal value in an expression
type LiteralValue struct {
	Value interface{}
	Type  AttributeDataType
}

func (l *LiteralValue) String() string {
	switch v := l.Value.(type) {
	case string:
		return `"` + v + `"`
	default:
		return fmt.Sprintf("%v", l.Value)
	}
}

// PermissionModel represents a complete permission model
type PermissionModel struct {
	Entities map[string]*Entity
	Rules    map[string]*Rule  // Global rules indexed by name
	Source   string // Source file path
}

// NewPermissionModel creates a new permission model
func NewPermissionModel() *PermissionModel {
	return &PermissionModel{
		Entities: make(map[string]*Entity),
		Rules:    make(map[string]*Rule),
	}
}

// AddEntity adds an entity to the model
func (m *PermissionModel) AddEntity(entity *Entity) {
	m.Entities[entity.Name] = entity
	
	// Add entity-specific rules to global rules
	for _, rule := range entity.Rules {
		m.Rules[rule.Name] = &rule
	}
}

// AddRule adds a global rule to the model
func (m *PermissionModel) AddRule(rule *Rule) {
	m.Rules[rule.Name] = rule
}

// GetEntity gets an entity by name
func (m *PermissionModel) GetEntity(name string) *Entity {
	return m.Entities[name]
}

// GetRule gets a rule by name
func (m *PermissionModel) GetRule(name string) *Rule {
	return m.Rules[name]
}

// SyncRulesToDatabase saves all rules to the database
func (m *PermissionModel) SyncRulesToDatabase(db interface{}) error {
	// This is a placeholder. In a full implementation, this would insert/update
	// rule definitions in the database from the parsed model.
	// The actual implementation would depend on the database interface used.
	return nil
}
