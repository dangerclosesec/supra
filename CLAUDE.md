# SUPRA Project Guidelines

## Build & Test Commands
- Setup environment: `make setup`
- Set environment variables: `make env`
- Database migration: `make migrate`
- Generate mocks: `make mocks`
- Validate permissions: `make validate-perms`
- Run all tests: `go test ./...`
- Run specific test: `go test ./path/to/package -run TestName`
- Run test with verbose output: `go test -v ./path/to/package`
- Run test with coverage: `go test -cover ./path/to/package`

## Code Style Guidelines
- Follow standard Go formatting conventions
- Use meaningful variable/function names (camelCase for variables, PascalCase for exports)
- Group imports: standard library first, then third-party, then internal
- Error handling: always check errors and provide context
- Use mockgen for test mocks (`go generate ./internal/repository/mock_gen.go`)
- Document public APIs with meaningful comments
- Keep functions small and focused on a single responsibility
- Use context.Context for propagating deadlines, cancellation signals, and request-scoped values