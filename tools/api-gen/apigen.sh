#!/bin/bash

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

GIT_REF_TAG=$(git describe --tags)
BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=$GIT_REF_TAG"

# To generate with the current commit version:
# go run -ldflags="$BUILD_LD_FLAGS" ./main.go "$@"

# During development the version is unset, therefore 0 to not commit a new api client each time.
go run $SCRIPTPATH/main.go "$@"