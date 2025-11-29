.PHONY: help
help: ## Show this help message
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@echo "Building..."
	@go build -ldflags="-X 'main.Version=$(shell git describe --tags --always --dirty)'" -o bin/aosd ./cmd/aos/main.go

.PHONY: run
run: ## Run the application
	@go run ./cmd/aos/main.go

.PHONY: test
test: ## Run unit tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# -------------------------------------------------------------
# Diagram generation (Mermaid via Docker)
# -------------------------------------------------------------
.PHONY: diagram diagram-svg diagram-png

diagram: diagram-svg diagram-png ## Generate both SVG and PNG diagrams

diagram-svg: ## Generate architecture.svg from docs/architecture.mmd
	@mkdir -p docs
	@docker run --rm -u $$(id -u):$$(id -g) \
	  -v $$PWD:/work ghcr.io/mermaid-js/mermaid-cli:10.9.0 \
	  -i /work/docs/architecture.mmd -o /work/docs/architecture.svg
	@echo "Generated docs/architecture.svg"

diagram-png: ## Generate architecture.png from docs/architecture.mmd
	@mkdir -p docs
	@docker run --rm -u $$(id -u):$$(id -g) \
	  -v $$PWD:/work ghcr.io/mermaid-js/mermaid-cli:10.9.0 \
	  -i /work/docs/architecture.mmd -o /work/docs/architecture.png
	@echo "Generated docs/architecture.png"

.PHONY: install
install: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download

.PHONY: verify
verify: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "All checks passed!"

.PHONY: ci
ci: verify ## Run CI pipeline locally
	@echo "CI pipeline completed successfully!"

.DEFAULT_GOAL := help
