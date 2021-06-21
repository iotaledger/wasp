all: build

build:
	go build ./...

test: install
	go clean -testcache
	go test ./... -timeout 20m

test-short:
	go clean -testcache
	go test --short ./...

install:
	go install -tags rocksdb ./...


.PHONY: all build test test-short

