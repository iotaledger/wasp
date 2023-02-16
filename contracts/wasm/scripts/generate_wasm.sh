#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm

GIT_REF_TAG=$(git describe --tags)
BUILD_LD_FLAGS="-X=github.com/iotaledger/wasp/core/app.Version=${GIT_REF_TAG}"
go install -ldflags ${BUILD_LD_FLAGS} $root_path/tools/schema

cd $contracts_path
bash scripts/schema_all.sh
