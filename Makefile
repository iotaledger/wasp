all: build

build:
	go build -tags rocksdb ./...

test: install
	go clean -testcache
	go test -tags rocksdb ./... -timeout 20m

test-short:
	go clean -testcache
	go test -tags rocksdb --short ./...

install:
	go install -tags rocksdb ./...

lint:
	golangci-lint run

.PHONY: all build test test-short

