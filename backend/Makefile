# Variables
BINARY_NAME=api-bin
DOCKER_IMAGE=price-paid-api
SOURCE_DIR=.

.PHONY: all build clean test run static

all: build

# Standard build
build:
	go build -o $(BINARY_NAME) $(SOURCE_DIR)/main.go

# 100% Truly Static Binary
# Note: Requires cgo for sqlite3, and -extldflags "-static" for full portability
static:
	CGO_ENABLED=1 go build \
		-ldflags="-s -w -extldflags '-static'" \
		-o $(BINARY_NAME) \
		$(SOURCE_DIR)/main.go

# Run the app
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f api-linux
	rm -f api-win.exe

# Cross-compile for Linux (Static)
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -extldflags '-static'" \
		-o api-linux \
		$(SOURCE_DIR)/main.go
