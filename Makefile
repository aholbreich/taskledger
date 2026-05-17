# TaskLedger Makefile — adapted from aholbreich/adr-tool
BINARY_NAME := tl
INSTALL_DIR ?= $(HOME)/bin
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
COUNT := $(shell git rev-list $(VERSION)..HEAD --count 2>/dev/null || echo 0)
LDFLAGS := -X main.version=$(VERSION)-$(COUNT)-$(COMMIT_HASH)

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build test bdd install rpm get-version bump amend dists clean cleancache \
	binary_linux_amd64 binary_linux_arm64 binary_darwin_amd64 binary_darwin_arm64 \
	binary_windows_amd64 binary_windows_arm64

# Build the local binary with version stamping.
build:
	go fmt ./...
	go mod tidy
	go build -o $(BINARY_NAME) -ldflags "$(LDFLAGS)"

# Run all tests (Go unit tests + godog BDD suite under ./bdd/...).
test:
	go test -v ./...

# Run only the BDD suite.
bdd:
	go test -v ./bdd/...

# Install the binary to $(INSTALL_DIR) (default $(HOME)/bin).
install: test build
	mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"


get-version:
	@echo $(VERSION)-$(COUNT)-$(COMMIT_HASH)

# Interactive tag bump. Prompts for a new version, tags HEAD, pushes the tag.
bump:
	@echo "Current version: $(VERSION)"
	@echo "Current commit hash: $(COMMIT_HASH)"
	@echo "Current count: $(COUNT)"
	@echo "Enter new version: "
	@read new_version; \
	git tag $$new_version; \
	git push origin $$new_version

# Amend the last commit with all current changes and force-push. Use with care.
amend:
	git add .
	git commit --amend --no-edit
	git push --force

# Cross-platform release archives.
binary_linux_amd64:
	mkdir -p build/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o build/linux-amd64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/linux-amd64 -czvf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)
	rm -rf build/linux-amd64

binary_linux_arm64:
	mkdir -p build/linux-arm64
	GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/linux-arm64 -czvf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)
	rm -rf build/linux-arm64

binary_darwin_amd64:
	mkdir -p build/darwin-amd64
	GOOS=darwin GOARCH=amd64 go build -o build/darwin-amd64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/darwin-amd64 -czvf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)
	rm -rf build/darwin-amd64

binary_darwin_arm64:
	mkdir -p build/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -o build/darwin-arm64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/darwin-arm64 -czvf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)
	rm -rf build/darwin-arm64

binary_windows_amd64:
	mkdir -p build/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o build/windows-amd64/$(BINARY_NAME).exe -ldflags "$(LDFLAGS)"
	zip -j $(BINARY_NAME)-windows-amd64.zip build/windows-amd64/$(BINARY_NAME).exe
	rm -rf build/windows-amd64

binary_windows_arm64:
	mkdir -p build/windows-arm64
	GOOS=windows GOARCH=arm64 go build -o build/windows-arm64/$(BINARY_NAME).exe -ldflags "$(LDFLAGS)"
	zip -j $(BINARY_NAME)-windows-arm64.zip build/windows-arm64/$(BINARY_NAME).exe
	rm -rf build/windows-arm64

dists: binary_linux_amd64 binary_linux_arm64 binary_darwin_amd64 binary_darwin_arm64 binary_windows_amd64 binary_windows_arm64

# Remove local build artifacts.
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-*
	rm -rf build/

# Also clear Go build and module caches (slow; rarely needed).
cleancache: clean
	go clean -cache -testcache -modcache
