# Variables
DIST_DIRECTORY := bin
BINARY := main
SOURCES := $(wildcard src/*.go)

build:
	go build -o $(DIST_DIRECTORY)/$(BINARY) $(SOURCES)

# Build for Linux
build-linux: GOOS := linux
build-linux: EXTENSION := 
build-linux: build-cross

# Build for Windows
build-windows: GOOS := windows
build-windows: EXTENSION := .exe
build-windows: build-cross

# Cross-compile for a specific platform (used by build-linux and build-windows)
build-cross: export GOOS := $(GOOS)
build-cross: BINARY := $(DIST_DIRECTORY)/$(BINARY)$(EXTENSION)
build-cross: $(DIST_DIRECTORY)/$(BINARY)

# Clean up generated files
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY).exe