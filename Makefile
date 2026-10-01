# The Answer Protocol (TAP): build tool.
# Run every target from the repository root: the server reads data/world.json
# and writes saves/ relative to the current directory.

GO      ?= go
BIN_DIR ?= bin
ADDR    ?= 127.0.0.1:4242

SERVER_BIN := $(BIN_DIR)/tap-server
CLI_BIN    := $(BIN_DIR)/tap-cli
GUI_BIN    := $(BIN_DIR)/tap-gui

.DEFAULT_GOAL := help
.PHONY: help install build run-server run-client run-client-gui lint test clean

help: ## Show this help
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables: ADDR=host:port (client target, default $(ADDR)), TAP_LOG_FILE=path, TAP_LOG_LEVEL=debug|info|warn|error"

install: ## Download the Go module dependencies
	$(GO) mod download

build: ## Compile the server, CLI client and GUI client into bin/
	$(GO) build -o $(SERVER_BIN) ./cmd/server
	$(GO) build -o $(CLI_BIN) ./cmd/cli
	$(GO) build -o $(GUI_BIN) ./cmd/gui

run-server: ## Run the server on :4242 (JSON logs on stderr)
	$(GO) run ./cmd/server

run-client: ## Run the CLI client (ADDR=host:port to change the server)
	$(GO) run ./cmd/cli $(ADDR)

run-client-gui: ## Run the GUI client
	LANGUAGE=$$(printf '%s' "$$LANGUAGE" | sed 's/^:*//;s/:*$$//;s/::*/:/g') $(GO) run ./cmd/gui

lint: ## Check formatting (gofmt) and run go vet
	@unformatted="$$(gofmt -l cmd)"; \
	if [ -n "$$unformatted" ]; then echo "gofmt needed for:"; echo "$$unformatted"; exit 1; fi
	$(GO) vet ./...

test: ## Run all automated tests
	$(GO) test ./...

clean: ## Remove build outputs (saves/ is kept)
	rm -rf $(BIN_DIR)
	$(GO) clean ./...
