.DEFAULT_GOAL := build

BUILD_PATH := ./cmd/*
BINARY_PATH := ./bin/matterfeed

COVERAGE_FILE := ./cover.out

.PHONY: setup tidy lint test build run clean vuln help

HOMEBREW = https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
PACKAGES = go gopls golangci-lint govulncheck pre-commit

# setup the development dependencies
tools:
	@echo "Setting up development environment..."
	@which brew >/dev/null || /bin/bash -c "$(curl -fsSL $(HOMEBREW))"
	@brew update
	@for pkg in $(PACKAGES); do \
		if ! brew list | grep -q $$pkg; then \
			echo "Installing $$pkg"; \
			brew install $$pkg --force; \
		else \
			echo "$$pkg is already installed."; \
		fi; \
	done
	@go install github.com/vladopajic/go-test-coverage/v2@latest
	@echo "Setup complete!"

# tidy for managing project dependencies
tidy:
	go mod tidy

# golangci-lint for comprehensive linting, with automatic fixes where applicable
lint: tidy
	golangci-lint run --fix

# run tests
test: lint
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	$$(go env GOPATH)/bin/go-test-coverage --config=./.testcoverage.yml

# build the project and place the binary in the bin directory
build: test
	go build -o $(BINARY_PATH) $(BUILD_PATH)

# run the binary
run: build
	$(BINARY_PATH)

# remove the binary
clean:
	rm -f $(BINARY_PATH) $(COVERAGE_FILE)

# check for vulnerabilities
vuln:
	@echo "Running govulncheck..."
	govulncheck ./...

# help with the commands
help:
	@echo "Makefile commands:"
	@echo "  make brew     - Ensure the software dependencies for development"
	@echo "  make build     - Build the binary"
	@echo "  make run       - Run the application"
	@echo "  make clean     - Remove the binary"
	@echo "  make lint      - Run linters with --fix flag for automatic fixes"
	@echo "  make test      - Run tests"
	@echo "  make vuln      - Check for vulnerabilities in dependencies"
