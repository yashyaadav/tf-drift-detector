BINARY  := tfdrift
PKG     := ./...
BIN_DIR := bin
GO      ?= go

GOTEST_FLAGS ?=
ifdef UPDATE
GOTEST_FLAGS += -update
endif

.PHONY: all build test golden fmt vet lint tidy check run clean

all: check build

build: ## Compile the tfdrift binary into ./bin
	$(GO) build -o $(BIN_DIR)/$(BINARY) ./cmd/tfdrift

test: ## Run tests (UPDATE=1 regenerates golden files)
	$(GO) test $(GOTEST_FLAGS) $(PKG)

golden: ## Regenerate golden test fixtures
	$(GO) test $(PKG) -update

fmt:
	$(GO) fmt $(PKG)

vet:
	$(GO) vet $(PKG)

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

tidy:
	$(GO) mod tidy

check: fmt vet lint test ## Full local gate

run: build ## Build and run (pass args via ARGS="scan --state ...")
	./$(BIN_DIR)/$(BINARY) $(ARGS)

clean:
	rm -rf $(BIN_DIR) coverage.out
