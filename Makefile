.PHONY: help install-lint lint test test-race vuln ci clean

GOLANGCI_LINT_VERSION := latest

install-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint:
	@golangci-lint run

test:
	@go test -v ./...

test-race:
	@go test -race -v ./...

vuln:
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

ci: test-race vuln lint
	@echo "✅ All CI checks passed"

clean:
	@go clean -testcache
	@rm -rf dist/
