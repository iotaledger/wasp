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
	    github.com/iotaledger/wasp/packages/testutil \
	    github.com/iotaledger/wasp/packages/peering \
	    github.com/iotaledger/wasp/packages/peering/tcp
test-short:
	go test ./packages/... ./plugins/... ./client/...

itest:
	go install
	go test github.com/iotaledger/wasp/tools/cluster/tests/wasptest_new

itest-by-one:
	go install
	go clean -testcache
	go test github.com/iotaledger/wasp/tools/cluster/tests/wasptest_new --list ".*" | grep "^Test" | xargs -tn1 go test github.com/iotaledger/wasp/tools/cluster/tests/wasptest_new --run

.PHONY: all build build-dkg test

