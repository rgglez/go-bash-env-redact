BINARY    := redactenv
SRC       := ./src
BUILD_DIR := build

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Output name: add .exe on Windows
ifeq ($(GOOS),windows)
  OUT := $(BUILD_DIR)/$(BINARY).exe
else
  OUT := $(BUILD_DIR)/$(BINARY)
endif

.PHONY: all build linux darwin windows clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUT) $(SRC)
	@echo "Built $(OUT)"

linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)_linux $(SRC)
	@echo "Built $(BUILD_DIR)/$(BINARY)_linux"

darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)_mac $(SRC)
	@echo "Built $(BUILD_DIR)/$(BINARY)_mac"

windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY).exe $(SRC)
	@echo "Built $(BUILD_DIR)/$(BINARY).exe"

all-platforms: linux darwin windows

clean:
	rm -rf $(BUILD_DIR)
