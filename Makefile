GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILD_FLAGS = rocksdb
LINKER_FLAGS = "-X github.com/iotaledger/wasp/packages/wasp.VersionHash=$(GIT_COMMIT)"

all: build-lint

build:
	go build -tags $(BUILD_FLAGS) -ldflags $(LINKER_FLAGS) ./...

build-lint: build lint

test: install
	go test -tags $(BUILD_FLAGS) ./... --timeout 30m --count 1 -failfast

test-short:
	go test -tags $(BUILD_FLAGS) --short --count 1 ./...

install:
	go install -tags $(BUILD_FLAGS) -ldflags $(LINKER_FLAGS) ./...

lint:
	golangci-lint run

gofumpt-list:
	gofumpt -l ./

.PHONY: all build test test-short lint gofumpt-list

