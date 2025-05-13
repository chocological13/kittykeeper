.PHONY: up
up:
	docker-compose up -d

.PHONY: down
down:
	docker-compose down

.PHONY: build
build:
	docker-compose build

.PHONY: ps
ps:
	docker-compose ps

.PHONY: run-user-service
run-user-service:
	cd services/user-service && make run

.PHONY: run-cat-service
run-cat-service:
	cd services/cat-service && make run

.PHONY: install-tools
install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: setup
setup: install-tools
	# ? Any additional setup steps if any here