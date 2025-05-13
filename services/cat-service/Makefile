include ../../.env
export

MIGRATIONS_DIR=internal/database/migrations

.PHONY: build
build:
	go build -o bin/user-service ./cmd/api

.PHONY: run
run:
	go run ./cmd/api

.PHONY: test
test:
	go test -v ./...

.PHONY: migrations-create
migrations-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrations-create name=your_migration_name"; \
		exit 1; \
	fi; \
	goose -dir $(MIGRATIONS_DIR) create "$(name)" sql

.PHONY: migrations-up
migrations-up:
	goose -dir $(MIGRATIONS_DIR) postgres "${CAT_DB_URL}" up
	@echo "🆙 Database migration(s) applied successfully!"

.PHONY: migrations-down
migrations-down:
	@if [ "$(all)" = "true" ]; then \
  		goose -dir $(MIGRATIONS_DIR) postgres "${CAT_DB_URL}" down; \
  		echo "🔃 Rolled back all migrations!"; \
  	elif [ -n "$(n)" ]; then \
  	  	goose -dir $(MIGRATIONS_DIR) postgres "${CAT_DB_URL}" down $(n); \
  	  	echo "🔃 Rolled back $(n) migration(s)!"; \
  	elif [ -n "$(v)" ]; then \
  	  	goose -dir $(MIGRATIONS_DIR) postgres "${CAT_DB_URL}" goto $(v); \
  	  	echo "🔃 Rolled back to version $(v)"; \
  	else \
		goose -dir $(MIGRATIONS_DIR) postgres "${CAT_DB_URL}" down 1; \
		echo "🔃 Rolled back 1 migration."; \
    fi

.PHONY: install-tools
install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest