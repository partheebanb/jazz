.PHONY: test coverage lint fmt build clean check help

# Run all tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run tests with coverage and open HTML report
coverage-html:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -s -w .

# Build binaries
build:
	go build -o bin/jazz-api ./main.go
	go build -o bin/jazz-migrate ./cmd/migrate/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Run all checks (format, lint, test)
check: fmt lint test

# Show help
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make coverage      - Run tests with coverage report"
	@echo "  make coverage-html - Run tests and open HTML coverage report"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make build         - Build binaries"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make check         - Run fmt + lint + test"
