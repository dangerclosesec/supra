package serializer_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/dangerclosesec/supra/internal/serializer"
)

var (
	serializers = make(serializer.Serializers)
)

type ExampleModal struct {
	Name  string `json:"name" szlr:"always"`
	Email string `json:"email" szlr:"scope:admin"`
	Phone string `json:"phone" szlr:"scope:admin,self"`
}

type ExampleSerializer struct{}

func (s *ExampleSerializer) Decode(input []byte, output any) error {
	// Decode the input into the output
	return nil
}

func (s *ExampleSerializer) Encode(input any, output io.ByteWriter) error {
	// Encode the input into the output
	return nil
}

func init() {
	serializer.Register(&ExampleModal{}, &ExampleSerializer{})
}

func TestRegister(t *testing.T) {
	// Register a model and its serializer
	serializer.Register(&ExampleModal{}, &ExampleSerializer{})
}

func TestSerializers(t *testing.T) {
	// Access the serializers map
	_ = serializers
}

func TestSerializer(t *testing.T) {
	// Create a new instance of a serializer
	serialize := &ExampleModal{
		Name:  "John Doe",
		Email: "johndo@mcstinkface.com",
		Phone: "123-456-7890",
	}

	// Encode the input into the output
	writer := bytes.NewBuffer([]byte{})
	if err := serializer.Encode(serialize, writer); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("writer: %v\n", writer.String())
	t.Log(writer.String())

	// Decode the input into the output
	// _ = serializer.Decode([]byte{}, ExampleModal{})
}
