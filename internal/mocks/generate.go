// internal/mocks/generate.go
package mocks

//go:generate go run github.com/vektra/mockery/v2 --all --keeptree --dir ../repository --output . --filename mock_${GOFILE}.go --packagename mocks
