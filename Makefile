# Dooprint Makefile
# Everything builds inside Docker: neither Go, Node nor Inno Setup has to be installed.

.PHONY: help version web linux windows installer package-linux dist checksums \
        test vet fmt fmt-check check ci clean \
        bump-version release release-patch release-minor release-major release-dry-run

.DEFAULT_GOAL := help

# Variables
VERSION      := $(shell cat VERSION)
MODULE       := github.com/apiservicesac/dooprint
LDFLAGS      := -s -w -X $(MODULE)/internal/app.Version=$(VERSION)
DIST         := dist
UID_GID      := $(shell id -u):$(shell id -g)

GO_IMAGE     := golang:1.25-bookworm
NODE_IMAGE   := node:22-bookworm
ISS_IMAGE    := amake/innosetup
LIBUSB       := 1.0.27

# Go module and build caches survive between runs.
GO_RUN       := docker run --rm -v "$(CURDIR)":/src -w /src \
                -v dooprint-gomod:/go/pkg/mod -v dooprint-gocache:/root/.cache/go-build $(GO_IMAGE)

LINUX_BIN    := $(DIST)/dooprint
WINDOWS_BIN  := $(DIST)/dooprint.exe
LINUX_PKG    := $(DIST)/dooprint-$(VERSION)-linux-amd64.tar.gz
INSTALLER    := $(DIST)/dooprint-$(VERSION)-windows-amd64-setup.exe

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[1;33m
BLUE   := \033[0;34m
RED    := \033[0;31m
NC     := \033[0m

help: ## Show this help message
	@printf "$(GREEN)Dooprint $(VERSION) - Build & Release Automation$(NC)\n"
	@printf "\n"
	@printf "$(BLUE)Available targets:$(NC)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-18s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

version: ## Show the current version
	@printf "$(VERSION)\n"

# Build targets
web: ## Build the web interface into internal/webui/dist
	@printf "$(GREEN)Building the web interface...$(NC)\n"
	docker run --rm -v "$(CURDIR)":/src -w /src/frontend $(NODE_IMAGE) bash -c "\
		npm ci --no-audit --no-fund --silent && npm run build && \
		rm -rf /src/internal/webui/dist && cp -r dist /src/internal/webui/dist && \
		chown -R $(UID_GID) node_modules dist /src/internal/webui/dist"

linux: ## Build the Linux binary (dist/dooprint)
	@printf "$(GREEN)Building $(LINUX_BIN) $(VERSION)...$(NC)\n"
	@mkdir -p $(DIST)
	$(GO_RUN) bash -c "\
		apt-get update -qq && apt-get install -y -qq libusb-1.0-0-dev pkg-config >/dev/null && \
		CGO_ENABLED=1 go build -buildvcs=false -trimpath -ldflags '$(LDFLAGS)' -o $(LINUX_BIN) ./cmd/dooprint && \
		chown $(UID_GID) $(LINUX_BIN)"

windows: ## Build the Windows binary (dist/dooprint.exe), libusb linked statically
	@printf "$(GREEN)Building $(WINDOWS_BIN) $(VERSION)...$(NC)\n"
	@mkdir -p $(DIST)
	$(GO_RUN) bash -c "\
		set -e; \
		apt-get update -qq && apt-get install -y -qq gcc-mingw-w64-x86-64 pkg-config curl make bzip2 >/dev/null; \
		curl -sL https://github.com/libusb/libusb/releases/download/v$(LIBUSB)/libusb-$(LIBUSB).tar.bz2 | tar -xj -C /tmp; \
		cd /tmp/libusb-$(LIBUSB) && ./configure -q --host=x86_64-w64-mingw32 --prefix=/opt/libusb-win --enable-static --disable-shared && make -s -j4 && make -s install; \
		cd /src && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc PKG_CONFIG_PATH=/opt/libusb-win/lib/pkgconfig \
		go build -buildvcs=false -trimpath -ldflags '$(LDFLAGS)' -o $(WINDOWS_BIN) ./cmd/dooprint; \
		chown $(UID_GID) $(WINDOWS_BIN)"

installer: ## Build the Windows installer from dist/dooprint.exe
	@printf "$(GREEN)Building $(INSTALLER)...$(NC)\n"
	@test -f $(WINDOWS_BIN) || (printf "$(RED)$(WINDOWS_BIN) is missing: run make windows first$(NC)\n" && exit 1)
	@rm -rf $(DIST)/.installer && mkdir -p $(DIST)/.installer && chmod 777 $(DIST)/.installer
	docker run --rm -v "$(CURDIR)":/work:ro -v "$(CURDIR)/$(DIST)/.installer":/out -w /work/packaging/windows $(ISS_IMAGE) \
		/Q /DAppVersion=$(VERSION) /DBinary=../../$(WINDOWS_BIN) /O/out dooprint.iss
	@cp $(DIST)/.installer/*.exe $(INSTALLER) && rm -rf $(DIST)/.installer

package-linux: ## Pack the Linux binary with its installer (tar.gz)
	@printf "$(GREEN)Packing $(LINUX_PKG)...$(NC)\n"
	@test -f $(LINUX_BIN) || (printf "$(RED)$(LINUX_BIN) is missing: run make linux first$(NC)\n" && exit 1)
	@rm -rf $(DIST)/.pkg && mkdir -p $(DIST)/.pkg/dooprint
	@cp $(LINUX_BIN) packaging/linux/install.sh packaging/linux/dooprint.service $(DIST)/.pkg/dooprint/
	@tar -czf $(LINUX_PKG) -C $(DIST)/.pkg dooprint && rm -rf $(DIST)/.pkg

checksums: ## Write SHA256SUMS for the release files in dist
	@cd $(DIST) && sha256sum dooprint-$(VERSION)-* > SHA256SUMS && cat SHA256SUMS

dist: web linux windows installer package-linux checksums ## Build every release file
	@printf "$(GREEN)Release files for $(VERSION):$(NC)\n"
	@ls -lh $(DIST)/dooprint-$(VERSION)-* $(DIST)/SHA256SUMS

# Quality targets
test: ## Run the Go tests
	@printf "$(GREEN)Running tests...$(NC)\n"
	$(GO_RUN) bash -c "apt-get update -qq && apt-get install -y -qq libusb-1.0-0-dev pkg-config >/dev/null && \
		mkdir -p internal/webui/dist && ( [ -e internal/webui/dist/index.html ] || touch internal/webui/dist/index.html ) && \
		go test ./internal/..."

vet: ## Run go vet
	@printf "$(GREEN)Running go vet...$(NC)\n"
	$(GO_RUN) bash -c "apt-get update -qq && apt-get install -y -qq libusb-1.0-0-dev pkg-config >/dev/null && \
		mkdir -p internal/webui/dist && ( [ -e internal/webui/dist/index.html ] || touch internal/webui/dist/index.html ) && \
		go vet ./..."

fmt: ## Format the Go code
	$(GO_RUN) bash -c "gofmt -w cmd internal && chown -R $(UID_GID) cmd internal"

fmt-check: ## Check the Go formatting without changing files
	@printf "$(GREEN)Checking formatting...$(NC)\n"
	@$(GO_RUN) bash -c 'files=$$(gofmt -l cmd internal); if [ -n "$$files" ]; then echo "Not formatted:"; echo "$$files"; exit 1; fi'
	@bash -n packaging/linux/install.sh

check: fmt-check vet ## Run code quality checks (format + vet)

ci: check test ## Run the CI pipeline locally

clean: ## Remove build outputs
	@printf "$(GREEN)Cleaning...$(NC)\n"
	rm -rf $(DIST) frontend/dist

# Version and release targets
bump-version: ## Set the version: make bump-version VERSION=1.2.3
	@./scripts/release.sh --bump-only $(VERSION)

release-patch: ## Release a patch version (X.Y.Z+1)
	@./scripts/release.sh patch

release-minor: ## Release a minor version (X.Y+1.0)
	@./scripts/release.sh minor

release-major: ## Release a major version (X+1.0.0)
	@./scripts/release.sh major

release: ## Release a given version: make release VERSION=1.2.3
	@./scripts/release.sh $(VERSION)

release-dry-run: ## Show what a patch release would do
	@./scripts/release.sh patch --dry-run
