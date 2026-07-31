-include .env
export

define require
@if [ -z "$($(1))" ]; then echo "$(1) is not set: add it to .env or pass $(1)=... "; exit 1; fi
endef

.PHONY: setup
setup:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
	go install gotest.tools/gotestsum@latest
	go install github.com/watchexec/watchexec@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

.PHONY: start
start:
	go run cmd/main.go

.PHONY: start-prod
start-prod:
	@set -a && [ -f .prod.env ] && . ./.prod.env; set +a; go run cmd/main.go

.PHONY: nabu
nabu:
	@PERSISTENCE_DIR=$(HOME)/Documents/nabu-persistence go run cmd/main.go

.PHONY: dev
dev:
	watchexec -e go -r make start

.PHONY: test
test:
	@gotestsum --format testname -- -v ./...

.PHONY: test-race
test-race:
	@gotestsum --format testname -- -race -v ./...

.PHONY: test-ci
test-ci:
	@gotestsum --format testname -- -race -coverprofile=coverage.out ./...

.PHONY: coverage
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

.PHONY: coverage-all
coverage-all:
	@go test -coverpkg=./... -coverprofile=coverage-all.out ./...
	@go tool cover -html=coverage-all.out

.PHONY: submit
submit:
	$(call require,PROJECT_ID)
	$(call require,JSON)
	@curl -X POST http://localhost:$(or $(PORT),8080)/commands/$(PROJECT_ID) \
		-H "Content-Type: application/json" \
		-d '$(JSON)'
