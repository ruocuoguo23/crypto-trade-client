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

# Generate binary names based on the directory structure
PLUGINS := $(addprefix $(BINARY_DIR)/,$(PLUGIN_FILES:cmd/%/main.go=%))

all: build

build: $(PLUGINS)

$(BINARY_DIR)/%: cmd/%/main.go
	$(GOBUILD) -o $@ $<


test:
	$(GOTEST) -v ./...


clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)


run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)