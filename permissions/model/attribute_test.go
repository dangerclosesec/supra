package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeValueParsing(t *testing.T) {
	tests := []struct {
		name     string
		dataType AttributeDataType
		input    string
		want     interface{}
		wantErr  bool
	}{
		{
			name:     "boolean true",
			dataType: AttributeTypeBoolean,
			input:    "true",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "boolean false",
			dataType: AttributeTypeBoolean,
			input:    "false",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "string",
			dataType: AttributeTypeString,
			input:    "hello world",
			want:     "hello world",
			wantErr:  false,
		},
		{
			name:     "integer",
			dataType: AttributeTypeInteger,
			input:    "42",
			want:     int64(42),
			wantErr:  false,
		},
		{
			name:     "double",
			dataType: AttributeTypeDouble,
			input:    "3.14159",
			want:     3.14159,
			wantErr:  false,
		},
		{
			name:     "string array",
			dataType: AttributeTypeStringArray,
			input:    `["one", "two", "three"]`,
			want:     []string{"one", "two", "three"},
			wantErr:  false,
		},
		{
			name:     "boolean array",
			dataType: AttributeTypeBooleanArray,
			input:    `[true, false, true]`,
			want:     []bool{true, false, true},
			wantErr:  false,
		},
		{
			name:     "integer array",
			dataType: AttributeTypeIntegerArray,
			input:    `[1, 2, 3]`,
			want:     []int64{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "double array",
			dataType: AttributeTypeDoubleArray,
			input:    `[1.1, 2.2, 3.3]`,
			want:     []float64{1.1, 2.2, 3.3},
			wantErr:  false,
		},
		{
			name:     "invalid boolean",
			dataType: AttributeTypeBoolean,
			input:    "not a boolean",
			want:     false,
			wantErr:  false, // This doesn't error because we just check if it's lowercase "true"
		},
		{
			name:     "invalid integer",
			dataType: AttributeTypeInteger,
			input:    "not an integer",
			wantErr:  true,
		},
		{
			name:     "invalid array",
			dataType: AttributeTypeStringArray,
			input:    "not an array",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAttributeValue(tt.dataType, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttributeValueValidation(t *testing.T) {
	tests := []struct {
		name     string
		dataType AttributeDataType
		value    interface{}
		wantErr  bool
	}{
		{
			name:     "valid boolean",
			dataType: AttributeTypeBoolean,
			value:    true,
			wantErr:  false,
		},
		{
			name:     "invalid boolean",
			dataType: AttributeTypeBoolean,
			value:    "true",
			wantErr:  true,
		},
		{
			name:     "valid string",
			dataType: AttributeTypeString,
			value:    "hello",
			wantErr:  false,
		},
		{
			name:     "invalid string",
			dataType: AttributeTypeString,
			value:    123,
			wantErr:  true,
		},
		{
			name:     "valid integer",
			dataType: AttributeTypeInteger,
			value:    int64(42),
			wantErr:  false,
		},
		{
			name:     "invalid integer",
			dataType: AttributeTypeInteger,
			value:    "42",
			wantErr:  true,
		},
		{
			name:     "valid double",
			dataType: AttributeTypeDouble,
			value:    3.14159,
			wantErr:  false,
		},
		{
			name:     "invalid double",
			dataType: AttributeTypeDouble,
			value:    "3.14159",
			wantErr:  true,
		},
		{
			name:     "valid string array",
			dataType: AttributeTypeStringArray,
			value:    []string{"one", "two", "three"},
			wantErr:  false,
		},
		{
			name:     "invalid string array",
			dataType: AttributeTypeStringArray,
			value:    []int{1, 2, 3},
			wantErr:  true,
		},
		{
			name:     "nil value",
			dataType: AttributeTypeString,
			value:    nil,
			wantErr:  true,
		},
		{
			name:     "unsupported data type",
			dataType: "unsupported",
			value:    "anything",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAttributeValue(tt.dataType, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidAttributeDataType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		want     bool
	}{
		{"boolean type", "boolean", true},
		{"boolean array type", "boolean[]", true},
		{"string type", "string", true},
		{"string array type", "string[]", true},
		{"integer type", "integer", true},
		{"integer array type", "integer[]", true},
		{"double type", "double", true},
		{"double array type", "double[]", true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidAttributeDataType(tt.dataType))
		})
	}
}