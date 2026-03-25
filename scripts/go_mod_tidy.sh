#!/bin/bash

GOFLAGS= go mod tidy

pushd tools/gendoc
GOFLAGS= go mod tidy
popd

pushd tools/wasp-cli
GOFLAGS= go mod tidy
popd

pushd tools/evm/evmemulator
GOFLAGS= go mod tidy
popd

# Re-vendor after tidy
go work vendor
./scripts/vendor_fix_cgo.sh
