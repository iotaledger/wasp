all: build-lint

build:
	go build -tags rocksdb ./...

build-lint: build lint

test: install
	go test -tags rocksdb ./... --timeout 30m --count 1 -failfast

test-short:
	go test -tags rocksdb --short --count 1 ./...

install:
	go install -tags rocksdb ./...

lint:
	golangci-lint run

gofumpt-list:
	gofumpt -l ./

.PHONY: all build test test-short lint gofumpt-list

