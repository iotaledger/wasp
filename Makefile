all: build

build:
	go build ./...

test: install
	go clean -testcache
	go test -tags rocksdb ./... -timeout 20m

test-short:
	go clean -testcache
	go test -tags rocksdb --short ./...

install:
	go install -tags rocksdb ./...


.PHONY: all build test test-short

