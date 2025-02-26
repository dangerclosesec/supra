// File: parser/parse.go
package parser

import (
	"io/ioutil"

	"github.com/dangerclosesec/supra/permissions/model"
)

// ParseFile parses a .perm file
func ParseFile(filePath string) (*model.PermissionModel, []string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, []string{err.Error()}, err
	}

	lexer := NewLexer(string(content))
	parser := NewParser(lexer)

	permModel := parser.ParsePermissionModel()
	permModel.Source = filePath

	return permModel, parser.Errors(), nil
}
