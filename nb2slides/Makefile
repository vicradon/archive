# Makefile for nb2slides

# Variables
INPUT ?= notebook.ipynb
OUTPUT_DIR ?= slides_output
PORT ?= 8080
MAX_LINES ?= 50
SHOW_OUTPUTS ?= true

# Default target
.PHONY: all
all: serve

# Serve HTML preview in browser
# Usage: make serve notebook.ipynb
#        make serve INPUT=notebook.ipynb
serve:
	@if [ -f "$(word 2,$(MAKECMDGOALS))" ]; then \
		INPUT_FILE=$(word 2,$(MAKECMDGOALS)); \
		echo "Starting preview server at http://localhost:$(PORT)..."; \
		echo "Input: $$INPUT_FILE"; \
		go run main.go -i "$$INPUT_FILE" -serve -port $(PORT); \
	elif [ -n "$(INPUT)" ]; then \
		echo "Starting preview server at http://localhost:$(PORT)..."; \
		echo "Input: $(INPUT)"; \
		go run main.go -i "$(INPUT)" -serve -port $(PORT); \
	else \
		echo "Error: No input file specified"; \
		echo "Usage: make serve <notebook.ipynb>"; \
		echo "   or: make serve INPUT=<notebook.ipynb>"; \
		exit 1; \
	fi

# Export to PDFs
# Usage: make export notebook.ipynb
#        make export INPUT=notebook.ipynb
.PHONY: export
export:
	@if [ -f "$(word 2,$(MAKECMDGOALS))" ]; then \
		INPUT_FILE=$(word 2,$(MAKECMDGOALS)); \
		echo "Exporting slides to $(OUTPUT_DIR)/..."; \
		echo "Input: $$INPUT_FILE"; \
		go run main.go -i "$$INPUT_FILE" -o $(OUTPUT_DIR) -max-lines $(MAX_LINES) -outputs=$(SHOW_OUTPUTS); \
	elif [ -n "$(INPUT)" ]; then \
		echo "Exporting slides to $(OUTPUT_DIR)/..."; \
		echo "Input: $(INPUT)"; \
		go run main.go -i "$(INPUT)" -o $(OUTPUT_DIR) -max-lines $(MAX_LINES) -outputs=$(SHOW_OUTPUTS); \
	else \
		echo "Error: No input file specified"; \
		echo "Usage: make export <notebook.ipynb>"; \
		echo "   or: make export INPUT=<notebook.ipynb>"; \
		exit 1; \
	fi

# Export without outputs
.PHONY: export-no-outputs
export-no-outputs:
	@echo "Exporting slides (without outputs) to $(OUTPUT_DIR)/..."
	go run main.go -i $(INPUT) -o $(OUTPUT_DIR) -max-lines $(MAX_LINES) -outputs=false

# Build binary
.PHONY: build
build:
	@echo "Building nb2slides..."
	go build -o nb2slides main.go

# Install to GOPATH/bin
.PHONY: install
install:
	@echo "Installing nb2slides..."
	go install

# Clean output directory
.PHONY: clean
clean:
	@echo "Cleaning $(OUTPUT_DIR)/..."
	rm -rf $(OUTPUT_DIR)

# Clean build artifacts
.PHONY: clean-all
clean-all: clean
	@echo "Cleaning build artifacts..."
	rm -f nb2slides

# Run tests
.PHONY: test
test:
	go test ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run ./...

# Show help
.PHONY: help
help:
	@echo "nb2slides - Jupyter Notebook to PDF Slides Converter"
	@echo ""
	@echo "Usage:"
	@echo "  make serve                    [INPUT=notebook.ipynb] [PORT=8080]"
	@echo "  make export                   [INPUT=notebook.ipynb] [OUTPUT_DIR=slides_output] [MAX_LINES=50] [SHOW_OUTPUTS=true]"
	@echo "  make export-no-outputs        [INPUT=notebook.ipynb] [OUTPUT_DIR=slides_output] [MAX_LINES=50]"
	@echo "  make build"
	@echo "  make install"
	@echo "  make clean"
	@echo "  make test"
	@echo ""
	@echo "Examples:"
	@echo "  make serve INPUT=../shape_detection.ipynb PORT=8080"
	@echo "  make export INPUT=../my_notebook.ipynb OUTPUT_DIR=my_slides"
	@echo "  make export INPUT=../large_notebook.ipynb MAX_LINES=30"
