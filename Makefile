GIT_REF_TAG := $(shell git describe --tags)
BUILD_TAGS = rocksdb
ifdef OS
# windows
BUILD_LD_FLAGS = "-X=github.com/iotaledger/wasp/components/app.Version=$(GIT_REF_TAG)"
else
ifeq ($(shell uname -m), arm64)
BUILD_LD_FLAGS = "-X=github.com/iotaledger/wasp/components/app.Version=$(GIT_REF_TAG) -extldflags \"-Wa,--noexecstack\""
else
BUILD_LD_FLAGS = "-X=github.com/iotaledger/wasp/components/app.Version=$(GIT_REF_TAG) -extldflags \"-z noexecstack\""
endif
endif
DOCKER_BUILD_ARGS = # E.g. make docker-build "DOCKER_BUILD_ARGS=--tag wasp:devel"

#
# You can override these e.g. as
#     make test TEST_PKG=./packages/vm/core/testcore/ TEST_ARG="-v --run TestAccessNodes"
#
TEST_PKG=./...
TEST_ARG=

BUILD_PKGS ?= ./ ./tools/cluster/wasp-cluster/
BUILD_CMD=go build -o . -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS)
INSTALL_CMD=go install -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS)
WASP_CLI_TAGS = no_wasmhost

# Docker image name and tag
DOCKER_IMAGE_NAME=wasp
DOCKER_IMAGE_TAG=develop

all: build-lint

wasm:
	bash contracts/wasm/scripts/schema_all.sh

compile-solidity:
	cd packages/vm/core/evm/iscmagic && go generate
	cd packages/evm/evmtest && go generate

build-cli:
	cd tools/wasp-cli && go mod tidy && go build -ldflags $(BUILD_LD_FLAGS) -tags ${WASP_CLI_TAGS} -o ../../

build-full: build-cli
	$(BUILD_CMD) ./...

build: build-cli
	$(BUILD_CMD) $(BUILD_PKGS)

build-lint: build lint

gendoc:
	./scripts/gendoc.sh

test-full: install
	go test -tags $(BUILD_TAGS),runheavy -ldflags $(BUILD_LD_FLAGS) ./... --timeout 60m --count 1 -failfast

test: install
	go test -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS) $(TEST_PKG) --timeout 90m --count 1 -failfast  $(TEST_ARG)

test-short:
	go test -tags $(BUILD_TAGS) -ldflags $(BUILD_LD_FLAGS) --short --count 1 -failfast $(shell go list ./... | grep -v github.com/iotaledger/wasp/contracts/wasm)

install-cli:
	cd tools/wasp-cli && go mod tidy && go install -ldflags $(BUILD_LD_FLAGS)

install-full: install-cli
	$(INSTALL_CMD) ./...

install: install-cli install-pkgs

install-pkgs:
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

docker-build:
	DOCKER_BUILDKIT=1 docker build ${DOCKER_BUILD_ARGS} \
		--build-arg BUILD_TAGS=${BUILD_TAGS} \
		--build-arg BUILD_LD_FLAGS=${BUILD_LD_FLAGS} \
		--tag iotaledger/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
		.

docker-check-push-deps:
    ifndef DOCKER_USERNAME
	    $(error DOCKER_USERNAME is undefined)
    endif
    ifndef DOCKER_ACCESS_TOKEN
	    $(error DOCKER_ACCESS_TOKEN is undefined)
    endif

docker-push:
	echo "$(DOCKER_ACCESS_TOKEN)" | docker login --username $(DOCKER_USERNAME) --password-stdin
	docker tag iotaledger/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	docker push $(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

docker-build-push: docker-check-push-deps docker-build docker-push

deps-versions:
	@grep -n "====" packages/testutil/privtangle/privtangle.go | \
		awk -F ":" '{ print $$1 }' | \
		{ read from ; read to; awk -v s="$$from" -v e="$$to" 'NR>1*s&&NR<1*e' packages/testutil/privtangle/privtangle.go; }

.PHONY: all wasm compile-solidity build-cli build-full build build-lint test-full test test-short install-cli install-full install lint gofumpt-list docker-build deps-versions
