# tl Makefile — adapted from aholbreich/adr-tool
BINARY_NAME := tl
INSTALL_DIR ?= $(HOME)/bin
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
COUNT := $(shell git rev-list $(VERSION)..HEAD --count 2>/dev/null || echo 0)
LDFLAGS := -X main.version=$(VERSION)-$(COUNT)-$(COMMIT_HASH)

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build test bdd install rpm get-version changelog release amend dists clean cleancache \
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

# Build a local RPM package in dist/rpm/.
rpm:
	bash .github/scripts/build-rpm.sh

get-version:
	@echo $(VERSION)-$(COUNT)-$(COMMIT_HASH)

# Generate release notes from commits since the previous tag.
changelog:
	@if prev=$$(git describe --tags --abbrev=0 2>/dev/null); then \
		range="$$prev..HEAD"; \
	else \
		range="HEAD"; \
	fi; \
	git log --oneline "$$range" | awk ' \
		BEGIN { print "## What'"'"'s Changed"; print "" } \
		{ \
			msg = $$0; \
			sub(/^[0-9a-f]+[[:space:]]+/, "", msg); \
			lower = tolower(msg); \
			line = "  • " msg "\n"; \
			if (lower ~ /^(feat|feature)(\([^)]+\))?!?:/) features = features line; \
			else if (lower ~ /^fix(\([^)]+\))?!?:/) fixes = fixes line; \
			else if (lower ~ /^docs(\([^)]+\))?!?:/) docs = docs line; \
			else if (lower ~ /^refactor(\([^)]+\))?!?:/) refactors = refactors line; \
			else if (lower ~ /^chores?(\([^)]+\))?!?:/) maintenance = maintenance line; \
			else other = other line; \
		} \
		END { \
			section("### 🚀 Features", features); \
			section("### 🐛 Fixes", fixes); \
			section("### 📖 Documentation", docs); \
			section("### ♻️ Refactoring", refactors); \
			section("### 🔧 Maintenance", maintenance); \
			section("### Other Changes", other); \
			if (!printed) print "No changes since the last release."; \
		} \
		function section(title, body) { \
			if (body != "") { \
				if (printed) print ""; \
				print title; \
				printf "%s", body; \
				printed = 1; \
			} \
		}'

ifneq ($(filter release,$(MAKECMDGOALS)),)
ifneq ($(origin VERSION),command line)
$(error Usage: make release VERSION=x.y.z)
endif
endif

# Validate, tag HEAD, and push the tag. GitHub Actions builds archives and publishes the GitHub Release.
release:
	@set -e; \
	printf '%s\n' "$(VERSION)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$' || { \
		echo "VERSION must look like x.y.z, got '$(VERSION)'" >&2; \
		exit 2; \
	}; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree is not clean; commit or stash changes before releasing." >&2; \
		exit 1; \
	fi; \
	branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$branch" != "main" ]; then \
		echo "Releases must be cut from main; current branch is $$branch." >&2; \
		exit 1; \
	fi; \
	git fetch origin main --tags; \
	if git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null; then \
		echo "Tag $(VERSION) already exists locally." >&2; \
		exit 1; \
	fi; \
	if git ls-remote --exit-code --tags origin "refs/tags/$(VERSION)" >/dev/null 2>&1; then \
		echo "Tag $(VERSION) already exists on origin." >&2; \
		exit 1; \
	fi; \
	local_head=$$(git rev-parse HEAD); \
	remote_head=$$(git rev-parse origin/main); \
	if [ "$$local_head" != "$$remote_head" ]; then \
		echo "HEAD ($$local_head) is not pushed to origin/main ($$remote_head). Push main before releasing." >&2; \
		exit 1; \
	fi; \
	go test -v ./...; \
	git tag -a "$(VERSION)" -m "tl $(VERSION)"; \
	git push origin "$(VERSION)"; \
	echo "Pushed tag $(VERSION). GitHub Actions will build archives and publish the GitHub Release."

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
