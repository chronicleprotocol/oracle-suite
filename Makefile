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
