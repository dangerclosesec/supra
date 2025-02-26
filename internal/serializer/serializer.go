package serializer

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
)

var (
	serializers           = make(Serializers)
	contextualSerializers = make(ContextualSerializers)
)

type ContextualSerializers map[reflect.Type]ContextualSerializer

// ContextualSerializer extends the base Serializer interface to accept a context.
type ContextualSerializer interface {
	// DecodeWithContext decodes the input into the output, using context for scope or auth info.
	DecodeWithContext(ctx context.Context, input []byte, output any) error

	// EncodeWithContext encodes the input into the output, using context for scope or auth info.
	EncodeWithContext(ctx context.Context, input any, output io.ByteWriter) error
}

type Serializers map[reflect.Type]Serializer

// Serializer is the interface that wraps the basic serialization methods
type Serializer interface {

	// Decode decodes the input into the output
	Decode(input []byte, output any) error

	// Encode encodes the input into the output
	Encode(input any, output io.ByteWriter) error
}

// Register registers a model and its serializer
func Register(model any, serializer Serializer) {
	serializers[reflect.TypeOf(model)] = serializer
}

func Encode(model any, output io.ByteWriter) error {
	if serializer, ok := serializers[reflect.TypeOf(model)]; ok {
		return serializer.Encode(model, output)
	}

	return fmt.Errorf("no serializer found for model %T", model)
}

func Decode(model any, input []byte) error {
	if serializer, ok := serializers[reflect.TypeOf(model)]; ok {
		return serializer.Decode(input, model)
	}

	return fmt.Errorf("no serializer found for model %T", model)
}

// RegisterContextual registers a model with a context-aware serializer.
func RegisterContextual(model any, serializer ContextualSerializer) {
	contextualSerializers[reflect.TypeOf(model)] = serializer
}

// EncodeWithContext attempts to find a context-aware serializer first.
// If none is found, it falls back to a basic serializer (optional).
func EncodeWithContext(ctx context.Context, model any, output io.ByteWriter) error {
	t := reflect.TypeOf(model)
	if s, ok := contextualSerializers[t]; ok {
		return s.EncodeWithContext(ctx, model, output)
	}

	// Fall back to the basic serializer if desired
	if s, ok := serializers[t]; ok {
		return s.Encode(model, output)
	}

	return fmt.Errorf("no serializer found for model %T", model)
}

// DecodeWithContext attempts to find a context-aware serializer first.
func DecodeWithContext(ctx context.Context, model any, input []byte) error {
	t := reflect.TypeOf(model)
	if s, ok := contextualSerializers[t]; ok {
		return s.DecodeWithContext(ctx, input, model)
	}

	if s, ok := serializers[t]; ok {
		return s.Decode(input, model)
	}

	return fmt.Errorf("no serializer found for model %T", model)
}

// ParseScopes extracts the scope portion from the tag. Example: "scope:admin,self" -> ["admin", "self"]
func ParseScopes(tag string) []string {
	// The tag might look like: `szlr:"scope:admin,self"`
	// so we look for "scope:" and everything after that
	prefix := "scope:"
	idx := strings.Index(tag, prefix)
	if idx == -1 {
		// If there's no "scope:" portion but the tag is "always", handle that
		if tag == "always" {
			return []string{"always"}
		}
		return nil
	}

	scopes := strings.TrimPrefix(tag[idx:], prefix)
	scopes = strings.TrimSpace(scopes)
	return strings.Split(scopes, ",")
}

type ScopedUser interface {
	Role() []string
	Scopes() []string
	UserID() uuid.UUID
}

// canViewField is a helper that examines the `szlr` tag and decides if the current user can see the field.
func CanViewField(szlrTag string, user *model.User, input any) bool {
	// If the tag is empty, you might default to always visible or always hidden
	if szlrTag == "" {
		return true // or false, depending on your security requirements
	}

	// // Example: parse "scope:admin,self"
	// // For “admin” we’d check user.Role == "admin"
	// // For “self” we might check user.ID matches the ID on the struct, etc.
	// scopes := ParseScopes(szlrTag)

	// if user == nil {
	// 	// If user is not present in context, either disallow everything or fallback to public logic
	// 	return false
	// }

	// // Check each scope in the tag
	// for _, scope := range scopes {
	// 	switch scope {
	// 	case "admin":
	// 		if user.Role == "admin" {
	// 			return true
	// 		}
	// 	case "self":
	// 		// If “self” is relevant, check if the resource belongs to the user
	// 		// This might require verifying if `input.(SomeModel).OwnerID == user.ID` for instance
	// 		// Example placeholder:
	// 		// if m, ok := input.(*model.ExampleModal); ok {
	// 		//     return (m.OwnerID == user.ID)
	// 		// }
	// 	case "always":
	// 		return true
	// 	default:
	// 		// Additional roles or scopes as needed
	// 	}
	// }

	return false
}
