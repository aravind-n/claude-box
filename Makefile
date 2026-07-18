# Build the Claude sandbox image with Apple `container` or Docker.
# Override the engine with ENGINE=docker.

ENGINE ?= $(shell if command -v container >/dev/null 2>&1; then echo container; \
		  elif command -v docker >/dev/null 2>&1; then echo docker; fi)
ifeq ($(strip $(ENGINE)),)
	$(error No container engine found; install one or pass ENGINE=docker)
endif

IMAGE      ?= claude-sandbox
TAG        ?= latest
IMAGE_REF  := $(IMAGE):$(TAG)
DOCKERFILE ?= image/Dockerfile

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
.PHONY: build rebuild clean

build:
	$(ENGINE) build $(BUILD_ARGS) --tag $(IMAGE_REF) --file $(DOCKERFILE) image

rebuild:
	$(ENGINE) build --no-cache $(BUILD_ARGS) --tag $(IMAGE_REF) --file $(DOCKERFILE) image

clean:
	-$(RM_IMAGE) $(IMAGE_REF)
