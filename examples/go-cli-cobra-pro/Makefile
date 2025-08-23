APP={{.project_name}}
GOPATH_BIN:=$(shell go env GOPATH)/bin

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Install dev tools
	@command -v golangci-lint >/dev/null || (echo "Installing golangci-lint" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@command -v goreleaser >/dev/null || (echo "Installing goreleaser" && go install github.com/goreleaser/goreleaser/v2@latest)

.PHONY: lint
lint: ## Run linters
	$(GOPATH_BIN)/golangci-lint run

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: build
build: ## Build binary
	go build -o $(APP) .

.PHONY: release
release: ## Dry-run release with goreleaser
	goreleaser release --clean --skip=publish --snapshot
