#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm

go install $root_path/tools/schema

cd $contracts_path
bash scripts/schema_all.sh
