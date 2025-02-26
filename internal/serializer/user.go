package serializer

import (
	"io"

	"github.com/dangerclosesec/supra/internal/model"
)

type UserSerializer struct{}

func (s *UserSerializer) Decode(input []byte, output any) error {
	// Decode the input into the output
	return nil
}

func (s *UserSerializer) Encode(input any, output io.ByteWriter) error {
	// Encode the input into the output
	return nil
}

func init() {
	Register(model.User{}, &UserSerializer{})
}
