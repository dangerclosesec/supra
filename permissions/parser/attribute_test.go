package parser

import (
	"testing"

	"github.com/dangerclosesec/supra/permissions/model"
	"github.com/stretchr/testify/assert"
)

func TestAttributeParsing(t *testing.T) {
	input := `entity test {
		// String attribute
		attribute name string
		
		// Boolean attribute
		attribute active boolean
		
		// Integer attribute
		attribute count integer
		
		// Double attribute
		attribute score double
		
		// Array attributes
		attribute tags string[]
		attribute flags boolean[]
		attribute values integer[]
		attribute rates double[]
	}`

	l := NewLexer(input)
	p := NewParser(l)
	permModel := p.ParsePermissionModel()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Check that the entity was parsed
	entity, exists := permModel.Entities["test"]
	if !exists {
		t.Fatalf("Entity 'test' not found in parsed model")
	}

	// Check that all attributes were parsed correctly
	expectedAttributes := []struct {
		name     string
		dataType model.AttributeDataType
	}{
		{"name", model.AttributeTypeString},
		{"active", model.AttributeTypeBoolean},
		{"count", model.AttributeTypeInteger},
		{"score", model.AttributeTypeDouble},
		{"tags", model.AttributeTypeStringArray},
		{"flags", model.AttributeTypeBooleanArray},
		{"values", model.AttributeTypeIntegerArray},
		{"rates", model.AttributeTypeDoubleArray},
	}

	assert.Equal(t, len(expectedAttributes), len(entity.Attributes), "Should have parsed all attributes")

	for i, expected := range expectedAttributes {
		if i < len(entity.Attributes) {
			attr := entity.Attributes[i]
			assert.Equal(t, expected.name, attr.Name, "Attribute name should match")
			assert.Equal(t, expected.dataType, attr.DataType, "Attribute data type should match")
		}
	}
}