# Makefile for a Go project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BINARY_DIR=bin

# Find all main.go files in the cmd directory
PLUGIN_FILES := $(shell find cmd -name 'main.go')

# Extract the relative paths without 'cmd/' and 'main.go'
PLUGIN_REL_PATHS := $(PLUGIN_FILES:cmd/%/main.go=%)
$(info PLUGIN_REL_PATHS: $(PLUGIN_REL_PATHS))

# Add the binary directory prefix
PLUGINS := $(addprefix $(BINARY_DIR)/,$(PLUGIN_REL_PATHS))
$(info PLUGINS: $(PLUGINS))

all: build

build: $(PLUGINS)

$(BINARY_DIR)/%: cmd/%/main.go
	@BINARY_NAME=$(subst /,-,$*); \
	echo "Building $< -> $(BINARY_DIR)/$$BINARY_NAME"; \
	echo "Source file: $<"; \
	$(GOBUILD) -o $(BINARY_DIR)/$$BINARY_NAME $<


test:
	$(GOTEST) -v ./...


clean:
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
