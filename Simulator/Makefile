# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=bin/IIoTOEE
BINARY_WINDOWS=$(BINARY_NAME)_windows_amd64.exe
BINARY_UNIX=$(BINARY_NAME)_unix_amd64
BINARY_DARWIN=$(BINARY_NAME)_darwin_amd64

all: build											## Build for all platforms

build: build-windows build-linux build-darwin		## Build binaries for all platforms

clean:												## Clean all binaries
	$(GOCLEAN)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_UNIX)

# Cross compilation
build-windows:										## Build binary for windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) -v

build-linux:										## Build binary for linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

build-darwin:										## Build binary for darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_DARWIN) -v

help: 												## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'