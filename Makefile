# Project structure
BIN_DIR = bin
CMD_DIR = cmd

# Targets
TARGETS = server create-user

# Build settings
GO = go
# Убираем -v чтобы не было лишнего вывода
GOFLAGS =
# Отключаем VCS информацию и сжимаем бинарник
LD_OPTS = -ldflags="-w -s" -buildvcs=false

.PHONY: all clean help $(TARGETS)

# Default target
.DEFAULT_GOAL := help

# Build all targets for current OS/Arch
all: $(TARGETS)

# Build server
server:
	@echo "Building server..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LD_OPTS) -o $(BIN_DIR)/server ./$(CMD_DIR)/server

# Build create-user tool
create-user:
	@echo "Building create-user..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LD_OPTS) -o $(BIN_DIR)/create-user ./$(CMD_DIR)/create-user

# Cross-compilation settings
PLATFORMS = windows linux darwin
ARCHS = amd64 arm64

# Cross-compile all targets
cross-all: $(addprefix cross-, $(TARGETS))

# Cross-compile server
cross-server:
	@echo "Cross-compiling server..."
	@for os in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			echo "  $$os/$$arch..."; \
			ext=; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(LD_OPTS) \
				-o $(BIN_DIR)/$$os/$$arch/server$$ext \
				./$(CMD_DIR)/server 2>/dev/null || echo "    (unsupported combination)"; \
		done; \
	done

# Cross-compile create-user
cross-create-user:
	@echo "Cross-compiling create-user..."
	@for os in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			echo "  $$os/$$arch..."; \
			ext=; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(LD_OPTS) \
				-o $(BIN_DIR)/$$os/$$arch/create-user$$ext \
				./$(CMD_DIR)/create-user 2>/dev/null || echo "    (unsupported combination)"; \
		done; \
	done

# Build for specific OS/Arch
build-%:
	$(eval PARTS := $(subst -, ,$@))
	$(eval GOOS := $(word 2,$(PARTS)))
	$(eval GOARCH := $(word 3,$(PARTS)))
	$(eval TARGET := $(word 4,$(PARTS)))
	@if [ -z "$(TARGET)" ]; then \
		echo "Usage: make build-<os>-<arch>-<target>"; \
		echo "Example: make build-linux-amd64-server"; \
		echo "Available targets: server, create-user"; \
		exit 1; \
	fi
	@echo "Building $(TARGET) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BIN_DIR)/$(GOOS)/$(GOARCH)
	@ext=; \
	if [ "$(GOOS)" = "windows" ]; then ext=".exe"; fi; \
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LD_OPTS) \
		-o $(BIN_DIR)/$(GOOS)/$(GOARCH)/$(TARGET)$$ext \
		./$(CMD_DIR)/$(TARGET)

# Quick builds for common platforms
build-windows:
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/server.exe ./$(CMD_DIR)/server
	GOOS=windows GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/create-user.exe ./$(CMD_DIR)/create-user

build-linux:
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/server ./$(CMD_DIR)/server
	GOOS=linux GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/create-user ./$(CMD_DIR)/create-user

build-macos:
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/server ./$(CMD_DIR)/server
	GOOS=darwin GOARCH=amd64 $(GO) build $(LD_OPTS) -o $(BIN_DIR)/create-user ./$(CMD_DIR)/create-user

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)

# Run tests
test:
	$(GO) test ./... -v

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Help message
help:
	@echo "Available commands:"
	@echo ""
	@echo "  Build for current OS:"
	@echo "    make all          - Build server and create-user"
	@echo "    make server       - Build server application"
	@echo "    make create-user  - Build create-user tool"
	@echo ""
	@echo "  Cross-compile for all platforms:"
	@echo "    make cross-all            - Build all for Windows/Linux/macOS (amd64/arm64)"
	@echo "    make cross-server         - Build server for all platforms"
	@echo "    make cross-create-user    - Build create-user for all platforms"
	@echo ""
	@echo "  Quick cross-compile:"
	@echo "    make build-windows        - Build all for Windows amd64"
	@echo "    make build-linux          - Build all for Linux amd64"
	@echo "    make build-macos          - Build all for macOS amd64"
	@echo ""
	@echo "  Specific target for specific platform:"
	@echo "    make build-<os>-<arch>-<target>"
	@echo "    Examples:"
	@echo "      make build-linux-arm64-server"
	@echo "      make build-darwin-amd64-create-user"
	@echo ""
	@echo "  Other commands:"
	@echo "    make clean        - Remove build artifacts"
	@echo "    make deps         - Download and tidy dependencies"
	@echo "    make test         - Run tests"
	@echo "    make help         - Show this help"