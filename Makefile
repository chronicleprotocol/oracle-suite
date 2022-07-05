PACKAGE ?= gofer
GO_FILES := $(shell { git ls-files; } | grep ".go$$")
LICENSED_FILES := $(shell { git ls-files; } | grep ".go$$")

BUILD_DIR := bin
BUILD_TARGET := $(BUILD_DIR)/gofer $(BUILD_DIR)/spire $(BUILD_DIR)/keeman
BUILD_FLAGS ?= all

OUT_DIR := workdir
COVER_FILE := $(OUT_DIR)/cover.out
TEST_FLAGS ?= all

GO := go

build: $(BUILD_TARGET)
.PHONY: build

$(BUILD_TARGET): export GOOS ?= linux
$(BUILD_TARGET): export GOARCH ?= amd64
$(BUILD_TARGET): export CGO_ENABLED ?= 0
$(BUILD_TARGET): $(GO_FILES)
	mkdir -p $(@D)
	$(GO) build -tags $(BUILD_FLAGS) $(LDFLAGS) -o $@ cmd/$(notdir $@)/*.go

clean:
	rm -rf $(OUT_DIR) $(BUILD_DIR)
.PHONY: clean

lint:
	golangci-lint run ./... --timeout 5m
.PHONY: lint

test:
	$(GO) test -v $$(go list ./... | grep -v /e2e/) -tags $(TEST_FLAGS)
.PHONY: test

test-api: export GOFER_TEST_API_CALLS = 1
test-api:
	$(GO) test ./pkg/origins/... -tags $(TEST_FLAGS) -testify.m TestRealAPICall
.PHONY: test-api

test-license: $(LICENSED_FILES)
	@grep -vlz "$$(tr '\n' . < LICENSE_HEADER)" $^ && exit 1 || exit 0
.PHONY: test-license

test-all: lint test test-license
.PHONY: test-all

cover:
	@mkdir -p $(dir $(COVER_FILE))
	$(GO) test -tags $(TEST_FLAGS) -coverprofile=$(COVER_FILE) ./...
	$(GO) tool cover -func=$(COVER_FILE)
.PHONY: cover

bench:
	$(GO) test -tags $(TEST_FLAGS) -bench=. ./...
.PHONY: bench

add-license: $(LICENSED_FILES)
	for x in $^; do tmp=$$(cat LICENSE_HEADER; sed -n '/^package \|^\/\/ *+build /,$$p' $$x); echo "$$tmp" > $$x; done
.PHONY: add-license

TEST_BUILD_TARGET := $(BUILD_DIR)/gofer-exchange.test
TEST_BUILD_PACKAGE := ./exchange
TEST_BUILD_PACKAGE_FILES := $(shell { git ls-files exchange; } | grep ".go$$")

build-test: $(TEST_BUILD_TARGET)
.PHONY: build-test

clean-test:
	rm $(TEST_BUILD_TARGET)
.PHONY: clean-test

$(TEST_BUILD_TARGET): clean-test $(TEST_BUILD_PACKAGE_FILES)
	mkdir -p $(@D)
	$(GO) test -tags $(TEST_FLAGS) -c -o $@ $(TEST_BUILD_PACKAGE)
.PHONY: build-test

run-test: $(TEST_BUILD_TARGET)
	$(TEST_BUILD_TARGET) -test.v -gofer.test-api-calls
.PHONY: run-test

WORMHOLE_DIR=e2e/wormhole
wormhole-e2e-up:
	@echo "Starting local testchain from snapshot."
	(cd $(WORMHOLE_DIR) && docker-compose up -d)
	@echo "Waiting for infra to start up fully..." && sleep 60 && docker ps -a && docker logs l2geth && docker logs deployer && echo "Ready."
	$(WORMHOLE_DIR)/auxiliary/scripts/wait-for-env.sh
.PHONY: wormhole-e2e-up

wormhole-e2e-down:
	@echo "Taking down local testchain"
	(cd $(WORMHOLE_DIR) && docker-compose down)
.PHONY: wormhole-e2e-down

wormhole-e2e:
	$(WORMHOLE_DIR)/e2e.sh
.PHONY: wormhole-e2e

# This specialized target will do everything in one shot, from bringing
# up/tearing down infra, building binaries, running E2E.
wormhole-e2e-one-shot: wormhole-e2e-down
	make wormhole-e2e-up
	make wormhole-e2e
	make wormhole-e2e-down
	make wormhole-e2e-clean
.PHONY: wormhole-e2e-one-shot

wormhole-e2e-clean:
	(cd $(WORMHOLE_DIR) && make clean)
.PHONY: wormhole-e2e-clean

wormhole-e2e-init-testchain:
	@echo "Building/running a local testchain from source..."
	@echo "Run this target once, just to set stuff up (takes a long time)."
	(cd $(WORMHOLE_DIR) && make init)
.PHONY: wormhole-e2e-init-testchain

wormhole-e2e-leeloo:
	./$(WORMHOLE_DIR)/start-leeloo.sh
.PHONY: wormhole-e2e-leeloo

wormhole-e2e-lair:
	./$(WORMHOLE_DIR)/start-lair.sh
.PHONY: wormhole-e2e-spire

wormhole-e2e-initiate:
	./$(WORMHOLE_DIR)/initiate-wormhole.sh $$(cat $(WORMHOLE_DIR)/auxiliary/optimism.json  | jq -r .l2_wormhole_gateway_address)
.PHONY: wormhole-e2e-initiate

VERSION_TAG_CURRENT := $(shell git tag --list 'v*' --points-at HEAD | sort --version-sort | tr \~ - | tail -1)
VERSION_TAG_LATEST := $(shell git tag --list 'v*' | tr - \~ | sort --version-sort | tr \~ - | tail -1)
ifeq ($(VERSION_TAG_CURRENT),$(VERSION_TAG_LATEST))
	VERSION := $(VERSION_TAG_CURRENT)
endif

VERSION_HASH := $(shell git rev-parse --short HEAD)
VERSION_DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
ifeq ($(VERSION),)
	VERSION := "dev-$(VERSION_HASH)-$(VERSION_DATE)"
endif

ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

LDFLAGS := -ldflags "-X github.com/chronicleprotocol/oracle-suite.Version=$(VERSION)"
