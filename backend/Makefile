.PHONY: help run build test test-integration bench bench-mem cover cover-html lint tidy mocks swagger seed-demo

## Default target: show help
help:
	@echo ""
	@echo "Available make targets:"
	@echo "  make run               - Run application (go run ./main.go)"
	@echo "  make build             - Build binary to ./bin/app"
	@echo "  make test              - Run all unit tests"
	@echo "  make test-integration  - Run integration tests (requires Docker)"
	@echo "  make bench             - Run all benchmarks"
	@echo "  make bench-mem         - Run all benchmarks with memory stats"
	@echo "  make cover             - Run tests with coverage in console"
	@echo "  make cover-html        - Run tests with coverage and open coverage.html"
	@echo "  make lint              - Run golangci-lint (if installed)"
	@echo "  make tidy              - go mod tidy"
	@echo "  make mocks             - Regenerate service mocks"
	@echo "  make swagger           - Generate swagger documentation"
	@echo "  make db-docs           - Generate database documentation from GORM models"
	@echo "  make seed-demo         - Seed realistic demo data (dev/demo only, refuses a non-empty DB)"
	@echo ""

run:
	go run ./main.go

build:
	mkdir -p bin
	go build -o bin/app .

test:
	go test ./...

test-integration:
	go test -v -tags=integration ./test/integration/...

bench:
	go test -bench=. -run=^# ./...

bench-mem:
	go test -bench=. -run=^# -benchmem ./...



cover:
	go test ./... -cover

cover-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

mocks:
	go generate ./...

swagger:
	swag init

db-docs:
	@echo "Generating database documentation..."
	@go run scripts/generate_db_docs.go -output DATABASE_AUTO.md
	@echo "✅ Documentation generated: DATABASE_AUTO.md"

seed-demo:
	go run ./cmd/seed


