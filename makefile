.PHONY: setup-hooks
setup:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
	go install gotest.tools/gotestsum@latest

.PHONY: test
test:
	@gotestsum --format testname -- -v ./...