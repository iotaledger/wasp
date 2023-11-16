#!/bin/bash

go mod tidy

pushd tools/gendoc
go mod tidy
popd

pushd tools/wasp-cli
go mod tidy
popd

pushd tools/gascalibration
go mod tidy
popd

pushd tools/evm/evmemulator
go mod tidy
popd
