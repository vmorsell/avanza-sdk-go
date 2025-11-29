.PHONY: install-lint lint test test-race vuln ci clean release-major release-minor release-patch

GOLANGCI_LINT_VERSION := v2.5.0

install-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint:
	@$(shell go env GOPATH)/bin/golangci-lint run

test:
	@go test -v ./...

test-race:
	@go test -race -v ./...

vuln:
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

ci: test-race vuln lint
	@echo "All CI checks passed"

clean:
	@go clean -testcache
	@rm -rf dist/

release-major:
	@./scripts/release.sh major

release-minor:
	@./scripts/release.sh minor

release-patch:
	@./scripts/release.sh patch