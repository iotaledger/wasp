#!/bin/bash

# determine current wasp version tag
GIT_REF_TAG="$(git describe --tags)"

# export variables for the build process
export BUILD_TAGS=rocksdb
export BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=${GIT_REF_TAG}"

# build the wasp container
docker compose build wasp