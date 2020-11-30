all: build

build:
	go build ./...

build-dkg:
	go build github.com/iotaledger/wasp/packages/dkg

test:
	go test ./...

test-kp:
	go clean -testcache && go test -timeout 30s \
	    github.com/iotaledger/wasp/packages/dkg \
	    github.com/iotaledger/wasp/packages/dks \
	    github.com/iotaledger/wasp/packages/testutil
test-short:
	go test ./packages/... ./plugins/... ./client/...

itest:
	go install
	go test github.com/iotaledger/wasp/tools/cluster/tests/wasptest_new

.PHONY: all build build-dkg test

