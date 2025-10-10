.PHONY: setup-hooks
setup-hooks:
	go install github.com/evilmartians/lefthook@latest
	lefthook install 