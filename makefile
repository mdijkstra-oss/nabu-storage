.PHONY: setup-hooks
setup:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
	go install gotest.tools/gotestsum@latest

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