package main

import (
	"fmt"
	"os"

	"github.com/dangerclosesec/supra/permissions/parser"
)

func main() {
	// Check if a schema file was provided
	schemaFile := "permissions/schema.perm"
	if len(os.Args) > 1 {
		schemaFile = os.Args[1]
	}

	// Read the schema file
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		fmt.Printf("Error reading schema file: %v\n", err)
		os.Exit(1)
	}

	// Parse the schema
	l := parser.NewLexer(string(content))
	p := parser.NewParser(l)
	permModel := p.ParsePermissionModel()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	// Print a summary of the parsed model
	fmt.Println("Successfully parsed permission model")
	fmt.Printf("Found %d entities\n", len(permModel.Entities))

	// Print details for each entity
	for name, entity := range permModel.Entities {
		fmt.Printf("\nEntity: %s\n", name)
		fmt.Printf("  Relations: %d\n", len(entity.Relations))
		fmt.Printf("  Permissions: %d\n", len(entity.Permissions))
		fmt.Printf("  Attributes: %d\n", len(entity.Attributes))

		// Print attribute details
		if len(entity.Attributes) > 0 {
			fmt.Println("  Attribute details:")
			for _, attr := range entity.Attributes {
				fmt.Printf("    - %s (%s)\n", attr.Name, attr.DataType)
			}
		}
	}
}