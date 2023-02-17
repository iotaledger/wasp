#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

cd tools/gendoc

# determine current wasp version tag
GIT_REF_TAG="$(git describe --tags)"

BUILD_TAGS=rocksdb
BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=${GIT_REF_TAG}"

go run -tags ${BUILD_TAGS} -ldflags ${BUILD_LD_FLAGS} main.go

cd ${CURRENT_DIR}

