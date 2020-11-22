all: build

build:
	go build ./...

build-dkg:
	go build github.com/iotaledger/wasp/packages/dkg

test:
	go test ./...

test-dkg:
	go clean -testcache && go test -v -timeout 30s github.com/iotaledger/wasp/packages/dkg

.PHONY: all build build-dkg test

