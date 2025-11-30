.DEFAULT_GOAL := help

GO ?= go
PKG ?= ./...
CMD_PKG ?= ./cmd

# Build tags used by `make run` / `make build`.
TAGS ?= fyne

# Build tags additionally verified by `make sanity` (set empty to skip tag verification).
# Note: on Windows, `fyne` often requires CGO + a C toolchain; use `make sanity-fyne` if available.
SANITY_TAGS ?=

# Optional app flags
# - OUT: passed as `-out <dir>` to the app
# - ARGS: extra args passed to the app (e.g. ARGS="-out D:\shots")
OUT ?=
ARGS ?=

TEST_FLAGS ?= -count=1

ifeq ($(OS),Windows_NT)
BIN_EXT := .exe
else
BIN_EXT :=
endif

BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/go-snip$(BIN_EXT)

ifeq ($(OS),Windows_NT)
MKDIR_BIN := powershell -NoProfile -NonInteractive -Command "New-Item -ItemType Directory -Force -Path '$(BIN_DIR)' | Out-Null"
else
MKDIR_BIN := mkdir -p "$(BIN_DIR)"
endif

.PHONY: help
help:
	@echo "go-snip Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make run            Run (defaults TAGS=fyne)"
	@echo "  make sanity         Format, tidy, vet, test, and build"
	@echo "  make sanity-fyne    Also verify fyne-tagged vet/test/build (requires CGO toolchain on Windows)"
	@echo "  make build          Build binary to $(BIN)"
	@echo "  make fmt            Run gofmt via 'go fmt'"
	@echo "  make tidy           Run 'go mod tidy'"
	@echo "  make vet            Run 'go vet' (no tags)"
	@echo "  make test           Run 'go test' (no tags)"
	@echo ""
	@echo "Common vars:"
	@echo "  TAGS=<tags>         Build tags for run/build (default: fyne). Use TAGS= to disable."
	@echo "  SANITY_TAGS=<tags>  Extra tag set verified by sanity (default: empty / skipped)."
	@echo "  OUT=<dir>           App output directory (passed as -out <dir>)"
	@echo "  ARGS=<args>         Extra args passed to app"

.PHONY: fmt
fmt:
	$(GO) fmt $(PKG)

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: vet
vet:
	$(GO) vet $(PKG)

.PHONY: test
test:
	$(GO) test $(TEST_FLAGS) $(PKG)

.PHONY: build
build:
	@$(MKDIR_BIN)
	$(GO) build $(if $(TAGS),-tags=$(TAGS),) -o "$(BIN)" $(CMD_PKG)

.PHONY: run
run:
	$(GO) run $(if $(TAGS),-tags=$(TAGS),) $(CMD_PKG) $(if $(OUT),-out "$(OUT)",) $(ARGS)

.PHONY: sanity sanity-notags sanity-tags vet-tags test-tags build-tags
sanity: sanity-notags sanity-tags

.PHONY: sanity-fyne
sanity-fyne:
	@$(MAKE) sanity SANITY_TAGS=fyne

sanity-notags: fmt tidy vet test
	@$(MKDIR_BIN)
	$(GO) build -o "$(BIN)" $(CMD_PKG)

ifeq ($(strip $(SANITY_TAGS)),)
sanity-tags:
	@echo "SANITY_TAGS is empty; skipping tag verification"
else
sanity-tags:
	@$(MKDIR_BIN)
	$(GO) vet -tags="$(SANITY_TAGS)" $(PKG)
	$(GO) test $(TEST_FLAGS) -tags="$(SANITY_TAGS)" $(PKG)
	$(GO) build -tags="$(SANITY_TAGS)" -o "$(BIN)" $(CMD_PKG)
endif

