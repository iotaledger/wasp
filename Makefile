all: build

build:
	go build ./...

build-dkg:
	go build github.com/iotaledger/wasp/packages/dkg

test:
	go test ./...

.PHONY: all build build-dkg test

