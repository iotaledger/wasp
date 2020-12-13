all: build

build:
	go build ./...

test:
	go install
	go clean -testcache
	go test ./... -timeout 20m

test-short:
	go clean -testcache
	go test --short ./...


.PHONY: all build test test-short

