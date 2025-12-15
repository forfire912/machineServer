.PHONY: all build run test clean

# Binary name
BINARY_NAME=machineserver

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: clean build

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	@echo "Running server..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf data/
	@rm -rf logs/

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run the server with custom config
run-config:
	@./$(BUILD_DIR)/$(BINARY_NAME) -config=$(CONFIG)
