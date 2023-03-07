GIT_REF_TAG := $(shell git describe --tags)
BUILD_TAGS = rocksdb
BUILD_LD_FLAGS = "-X=github.com/iotaledger/wasp/core/app.Version=$(GIT_REF_TAG)"
DOCKER_BUILD_ARGS = # E.g. make docker-build "DOCKER_BUILD_ARGS=--tag wasp:devel"

#
# You can override these e.g. as
#     make test TEST_PKG=./packages/vm/core/testcore/ TEST_ARG="-v --run TestAccessNodes"
#
TEST_PKG=./...
TEST_ARG=

BUILD_PKGS=./ ./tools/cluster/wasp-cluster/
BUILD_CMD=go build -o . -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS)
INSTALL_CMD=go install -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS)

all: build-lint

wasm:
	go install tools/schema
	cd contracts/wasm && schema -go -rs -ts

compile-solidity:
ifdef SKIP_SOLIDITY
	@echo "skipping compile-solidity rule"
else ifeq (, $(shell which solc))
	@echo "no solc found in PATH, evm contracts won't be compiled"
else
	cd packages/vm/core/evm/iscmagic && go generate
	cd packages/evm/evmtest && go generate
endif

build-cli:
	cd tools/wasp-cli && go mod tidy && go build -ldflags $(BUILD_LD_FLAGS) -o ../../

build-full: compile-solidity build-cli
	$(BUILD_CMD) ./...

build: compile-solidity build-cli
	$(BUILD_CMD) $(BUILD_PKGS)

build-lint: build lint

test-full: install
	go test -tags $(BUILD_TAGS),runheavy ./... --timeout 60m --count 1 -failfast -p 1

test: install
	go test -tags $(BUILD_TAGS) $(TEST_PKG) --timeout 90m --count 1 -failfast -p 1  $(TEST_ARG)

test-short:
	go test -tags $(BUILD_TAGS) --short --count 1 -failfast -p 1 $(shell go list ./... | grep -v github.com/iotaledger/wasp/contracts/wasm)

install-cli:
	cd tools/wasp-cli && go mod tidy && go install -ldflags $(BUILD_LD_FLAGS)

install-full: compile-solidity install-cli
	$(INSTALL_CMD) ./...

install: compile-solidity install-cli
	$(INSTALL_CMD) $(BUILD_PKGS)

lint: lint-wasp-cli
	golangci-lint run --timeout 5m

lint-wasp-cli:
	cd ./tools/wasp-cli && golangci-lint run --timeout 5m

apiclient:
	./clients/apiclient/generate_client.sh

apiclient-docker:
	./clients/apiclient/generate_client.sh docker

gofumpt-list:
	gofumpt -l ./

docker-build: compile-solidity
	DOCKER_BUILDKIT=1 docker build ${DOCKER_BUILD_ARGS} \
		--build-arg BUILD_TAGS=${BUILD_TAGS} \
		--build-arg BUILD_LD_FLAGS=${BUILD_LD_FLAGS} \
		.

deps-versions:
	@grep -n "====" packages/testutil/privtangle/privtangle.go | \
		awk -F ":" '{ print $$1 }' | \
		{ read from ; read to; awk -v s="$$from" -v e="$$to" 'NR>1*s&&NR<1*e' packages/testutil/privtangle/privtangle.go; }

.PHONY: all wasm compile-solidity build-cli build-full build build-lint test-full test test-short install-cli install-full install lint gofumpt-list docker-build deps-versions
