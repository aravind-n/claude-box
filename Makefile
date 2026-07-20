# Build the vhrn CLI (a static Go binary) and the container images it drives,
# with Apple `container` or Docker. Override the engine with ENGINE=docker.

ENGINE ?= $(shell if command -v container >/dev/null 2>&1; then echo container; \
		  elif command -v docker >/dev/null 2>&1; then echo docker; fi)
ifeq ($(strip $(ENGINE)),)
  $(error No container engine found; install one or pass ENGINE=docker)
endif

IMAGE      ?= vhrn-sandbox
TAG        ?= latest
IMAGE_REF  := $(IMAGE):$(TAG)
DOCKERFILE ?= image/Dockerfile

# The egress proxy sidecar image (see proxy/).
PROXY_IMAGE ?= vhrn-proxy
PROXY_REF   := $(PROXY_IMAGE):$(TAG)
PROXY_DIR   ?= proxy

# Where the CLI installs. PREFIX/BINDIR override the destination.
PREFIX   ?= /usr/local
BINDIR   ?= $(PREFIX)/bin
BIN_NAME ?= vhrn

# Go CLI build. Static (CGO off) so the binary is self-contained for curl-install.
GO        ?= go
PLATFORMS ?= darwin/arm64 darwin/amd64 linux/arm64 linux/amd64
DIST      ?= dist

# Match the container user to your host UID/GID (native Linux Docker only).
BUILD_ARGS :=
ifeq ($(ENGINE),docker)
  ifeq ($(shell uname -s),Linux)
    BUILD_ARGS := --build-arg USER_UID=$(shell id -u) --build-arg USER_GID=$(shell id -g)
  endif
endif

ifeq ($(ENGINE),docker)
RM_IMAGE := $(ENGINE) image rm
else
RM_IMAGE := $(ENGINE) image delete
endif

.DEFAULT_GOAL := build
.PHONY: build binary release test build-box build-proxy rebuild clean install uninstall

# Build both images the CLI needs: the box and its egress proxy sidecar.
build: build-box build-proxy

# The vhrn CLI: a single static binary (host arch).
binary:
	CGO_ENABLED=0 $(GO) build -o $(BIN_NAME) .

# Cross-compile release binaries into $(DIST) for curl-install distribution.
release:
	@mkdir -p $(DIST)
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; \
	  out=$(DIST)/$(BIN_NAME)-$$os-$$arch; \
	  echo "building $$out"; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build -o $$out . || exit 1; \
	done

# CLI + proxy unit tests.
test:
	$(GO) test ./...
	cd $(PROXY_DIR) && $(GO) test ./...

build-box:
	$(ENGINE) build $(BUILD_ARGS) --tag $(IMAGE_REF) --file $(DOCKERFILE) image

build-proxy:
	$(ENGINE) build --tag $(PROXY_REF) --file $(PROXY_DIR)/Dockerfile $(PROXY_DIR)

rebuild:
	$(ENGINE) build --no-cache $(BUILD_ARGS) --tag $(IMAGE_REF) --file $(DOCKERFILE) image
	$(ENGINE) build --no-cache --tag $(PROXY_REF) --file $(PROXY_DIR)/Dockerfile $(PROXY_DIR)

clean:
	-$(RM_IMAGE) $(IMAGE_REF)
	-$(RM_IMAGE) $(PROXY_REF)
	-rm -f $(BIN_NAME)
	-rm -rf $(DIST)

# Build the CLI and install it into $(BINDIR) (needs sudo for /usr/local/bin).
install: binary
	sudo install -m 0755 $(BIN_NAME) $(BINDIR)/$(BIN_NAME)
	@echo "Installed $(BINDIR)/$(BIN_NAME) — run '$(BIN_NAME)' in any project."

uninstall:
	sudo rm -f $(BINDIR)/$(BIN_NAME)
