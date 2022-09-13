#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm
cd $contracts_path
bash scripts/schema_all.sh
golangci-lint run --fix
cd $root_path/packages/wasmvm
golangci-lint run --fix
