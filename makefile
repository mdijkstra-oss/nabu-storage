include .env
export

.PHONY: setup
setup:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
	go install gotest.tools/gotestsum@latest
	go install github.com/watchexec/watchexec@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: start
start:
	go run cmd/main.go

.PHONY: start-prod
start-prod:
	@set -a && . ./.prod.env && set +a && go run cmd/main.go

.PHONY: nabu
nabu:
	@set -a && . ./.env && export PERSISTENCE_DIR=$(HOME)/Documents/nabu-persistence && go run cmd/main.go

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
	@curl -X POST http://localhost:$(PORT)/commands/$(PROJECT_ID) \
		-H "Content-Type: application/json" \
		-d '$(JSON)'