#!/bin/bash
cd ./contracts/wasm/scripts
bash schema_all.sh
cd ..
golangci-lint run --fix
cd ../../packages/wasmvm
golangci-lint run --fix
cd ../..
