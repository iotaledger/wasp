#!/bin/bash

GIT_REF_TAG=$(git describe --tags)
BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=$GIT_REF_TAG"

# To generate with the current commit version:
# go run -ldflags="$BUILD_LD_FLAGS" ./main.go "$@"

# During development the version is unset, therefore 0 to not commit a new api client each time.
go run ./main.go "$@"