package model

// Entity represents a permify entity type
type Entity struct {
	Name        string
	Relations   []Relation
	Permissions []Permission
	Comments    []string
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

// Parentheses represents a parenthesized expression
type Parentheses struct {
	Expr Expression
}

func (p *Parentheses) String() string {
	return "(" + p.Expr.String() + ")"
}

// PermissionModel represents a complete permission model
type PermissionModel struct {
	Entities map[string]*Entity
	Source   string // Source file path
}

// NewPermissionModel creates a new permission model
func NewPermissionModel() *PermissionModel {
	return &PermissionModel{
		Entities: make(map[string]*Entity),
	}
}

// AddEntity adds an entity to the model
func (m *PermissionModel) AddEntity(entity *Entity) {
	m.Entities[entity.Name] = entity
}

// GetEntity gets an entity by name
func (m *PermissionModel) GetEntity(name string) *Entity {
	return m.Entities[name]
}
