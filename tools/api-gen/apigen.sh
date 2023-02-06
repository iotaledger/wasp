#!/bin/bash

GIT_REF_TAG=$(git describe --tags)
BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=$GIT_REF_TAG"

go run -ldflags="$BUILD_LD_FLAGS" ./main.go "$@"