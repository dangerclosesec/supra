setup:
	go install go install go.uber.org/mock/mockgen@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	brew install permify/tap/permify

env:
	@bash -c 'set -a; . ./.env; set +a; exec $$SHELL'

migrate:
	@bash -c 'set -a; . ./.env; set +a; goose -dir db/migrations/$$DB_DRIVER $$DB_DRIVER "$$DB_URL" up'

migrate-down:
	@bash -c 'set -a; . ./.env; set +a; goose -dir db/migrations/$$DB_DRIVER $$DB_DRIVER "$$DB_URL" down-to 0'

mocks:
	go generate ./internal/repository/mock_gen.go  

validate-perms:
	permify validate permissions/validate.yml 