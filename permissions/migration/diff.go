// File: migration/diff.go
package migration

import (
	"fmt"
	"strings"

	"github.com/dangerclosesec/supra/permissions/model"
)

// ModelDiff represents the differences between two permission models
type ModelDiff struct {
	AddedEntities    []string
	RemovedEntities  []string
	ModifiedEntities map[string]*EntityDiff
	AddedRules       []model.Rule
	RemovedRules     []string
	ModifiedRules    map[string]*RuleDiff
}

// EntityDiff represents the differences between two entities
type EntityDiff struct {
	AddedPermissions    []model.Permission
	RemovedPermissions  []string
	ModifiedPermissions map[string]*PermissionDiff
}

// PermissionDiff represents the differences between two permissions
type PermissionDiff struct {
	OldExpression string
	NewExpression string
}

// RuleDiff represents the differences between two rules
type RuleDiff struct {
	OldExpression string
	NewExpression string
	Parameters    ParametersDiff
}

// ParametersDiff represents differences in rule parameters
type ParametersDiff struct {
	Added    []model.RuleParameter
	Removed  []string
	Modified map[string]struct{
		OldType string
		NewType string
	}
}

// GenerateDiff generates a diff between two permission models
func GenerateDiff(oldModel, newModel *model.PermissionModel) *ModelDiff {
	diff := &ModelDiff{
		ModifiedEntities: make(map[string]*EntityDiff),
		ModifiedRules:    make(map[string]*RuleDiff),
	}

	// Find added and removed entities
	oldEntities := make(map[string]bool)
	newEntities := make(map[string]bool)

	for name := range oldModel.Entities {
		oldEntities[name] = true
	}

	for name := range newModel.Entities {
		newEntities[name] = true
	}

	// Find added entities
	for name := range newEntities {
		if !oldEntities[name] {
			diff.AddedEntities = append(diff.AddedEntities, name)
		}
	}

	// Find removed entities
	for name := range oldEntities {
		if !newEntities[name] {
			diff.RemovedEntities = append(diff.RemovedEntities, name)
		}
	}

	// Compare common entities
	for name, newEntity := range newModel.Entities {
		oldEntity, exists := oldModel.Entities[name]
		if !exists {
			continue // Already handled as added entity
		}

		// Compare permissions
		entityDiff := compareEntityPermissions(oldEntity, newEntity)
		if !entityDiff.IsEmpty() {
			diff.ModifiedEntities[name] = entityDiff
		}
	}
	
	// Find added and removed rules
	oldRules := make(map[string]bool)
	newRules := make(map[string]bool)
	
	for name := range oldModel.Rules {
		oldRules[name] = true
	}
	
	for name := range newModel.Rules {
		newRules[name] = true
	}
	
	// Find added rules
	for name, rule := range newModel.Rules {
		if !oldRules[name] {
			diff.AddedRules = append(diff.AddedRules, *rule)
		}
	}
	
	// Find removed rules
	for name := range oldRules {
		if !newRules[name] {
			diff.RemovedRules = append(diff.RemovedRules, name)
		}
	}
	
	// Compare common rules
	for name, newRule := range newModel.Rules {
		oldRule, exists := oldModel.Rules[name]
		if !exists {
			continue // Already handled as added rule
		}
		
		// Compare rule expressions
		if oldRule.Expression != newRule.Expression {
			ruleDiff := &RuleDiff{
				OldExpression: oldRule.Expression,
				NewExpression: newRule.Expression,
			}
			diff.ModifiedRules[name] = ruleDiff
		}
	}

	return diff
}

// compareEntityPermissions compares the permissions of two entities
func compareEntityPermissions(oldEntity, newEntity *model.Entity) *EntityDiff {
	diff := &EntityDiff{
		ModifiedPermissions: make(map[string]*PermissionDiff),
	}

	oldPermissions := make(map[string]model.Permission)
	newPermissions := make(map[string]model.Permission)

	for _, perm := range oldEntity.Permissions {
		oldPermissions[perm.Name] = perm
	}

	for _, perm := range newEntity.Permissions {
		newPermissions[perm.Name] = perm
	}

	// Find added permissions
	for name, perm := range newPermissions {
		if _, exists := oldPermissions[name]; !exists {
			diff.AddedPermissions = append(diff.AddedPermissions, perm)
		}
	}

	// Find removed permissions
	for name := range oldPermissions {
		if _, exists := newPermissions[name]; !exists {
			diff.RemovedPermissions = append(diff.RemovedPermissions, name)
		}
	}

	// Compare common permissions
	for name, newPerm := range newPermissions {
		oldPerm, exists := oldPermissions[name]
		if !exists {
			continue // Already handled as added permission
		}

		// Compare expressions
		if oldPerm.Expression != newPerm.Expression {
			diff.ModifiedPermissions[name] = &PermissionDiff{
				OldExpression: oldPerm.Expression,
				NewExpression: newPerm.Expression,
			}
		}
	}

	return diff
}

// IsEmpty returns true if the diff is empty
func (d *EntityDiff) IsEmpty() bool {
	return len(d.AddedPermissions) == 0 &&
		len(d.RemovedPermissions) == 0 &&
		len(d.ModifiedPermissions) == 0
}

// String returns a string representation of the model diff
func (d *ModelDiff) String() string {
	var sb strings.Builder

	sb.WriteString("Permission Model Changes:\n\n")

	if len(d.AddedEntities) > 0 {
		sb.WriteString("Added Entities:\n")
		for _, entity := range d.AddedEntities {
			sb.WriteString(fmt.Sprintf("  + %s\n", entity))
		}
		sb.WriteString("\n")
	}

	if len(d.RemovedEntities) > 0 {
		sb.WriteString("Removed Entities:\n")
		for _, entity := range d.RemovedEntities {
			sb.WriteString(fmt.Sprintf("  - %s\n", entity))
		}
		sb.WriteString("\n")
	}

	if len(d.ModifiedEntities) > 0 {
		sb.WriteString("Modified Entities:\n")
		for entity, diff := range d.ModifiedEntities {
			sb.WriteString(fmt.Sprintf("  * %s:\n", entity))

			if len(diff.AddedPermissions) > 0 {
				sb.WriteString("    Added Permissions:\n")
				for _, perm := range diff.AddedPermissions {
					sb.WriteString(fmt.Sprintf("      + %s = %s\n", perm.Name, perm.Expression))
				}
			}

			if len(diff.RemovedPermissions) > 0 {
				sb.WriteString("    Removed Permissions:\n")
				for _, perm := range diff.RemovedPermissions {
					sb.WriteString(fmt.Sprintf("      - %s\n", perm))
				}
			}

			if len(diff.ModifiedPermissions) > 0 {
				sb.WriteString("    Modified Permissions:\n")
				for perm, permDiff := range diff.ModifiedPermissions {
					sb.WriteString(fmt.Sprintf("      * %s:\n", perm))
					sb.WriteString(fmt.Sprintf("        - %s\n", permDiff.OldExpression))
					sb.WriteString(fmt.Sprintf("        + %s\n", permDiff.NewExpression))
				}
			}
		}
	}
	
	// Added rules
	if len(d.AddedRules) > 0 {
		sb.WriteString("Added Rules:\n")
		for _, rule := range d.AddedRules {
			sb.WriteString(fmt.Sprintf("  + rule %s(", rule.Name))
			
			// Format parameters
			for i, param := range rule.Parameters {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s %s", param.Name, param.DataType))
			}
			sb.WriteString(fmt.Sprintf(") = %s\n", rule.Expression))
		}
		sb.WriteString("\n")
	}
	
	// Removed rules
	if len(d.RemovedRules) > 0 {
		sb.WriteString("Removed Rules:\n")
		for _, ruleName := range d.RemovedRules {
			sb.WriteString(fmt.Sprintf("  - %s\n", ruleName))
		}
		sb.WriteString("\n")
	}
	
	// Modified rules
	if len(d.ModifiedRules) > 0 {
		sb.WriteString("Modified Rules:\n")
		for ruleName, ruleDiff := range d.ModifiedRules {
			sb.WriteString(fmt.Sprintf("  * %s:\n", ruleName))
			sb.WriteString(fmt.Sprintf("    - %s\n", ruleDiff.OldExpression))
			sb.WriteString(fmt.Sprintf("    + %s\n", ruleDiff.NewExpression))
		}
		sb.WriteString("\n")
	}

	if len(d.AddedEntities) == 0 && 
	   len(d.RemovedEntities) == 0 && 
	   len(d.ModifiedEntities) == 0 &&
	   len(d.AddedRules) == 0 &&
	   len(d.RemovedRules) == 0 &&
	   len(d.ModifiedRules) == 0 {
		sb.WriteString("No changes detected.\n")
	}

	return sb.String()
}

// IsEmpty returns true if the diff contains no changes
func (d *ModelDiff) IsEmpty() bool {
	// Check if there are any entity additions, removals, modifications,
	// or rule additions, removals, modifications
	return len(d.AddedEntities) == 0 &&
		len(d.RemovedEntities) == 0 &&
		len(d.ModifiedEntities) == 0 &&
		len(d.AddedRules) == 0 &&
		len(d.RemovedRules) == 0 &&
		len(d.ModifiedRules) == 0
}
