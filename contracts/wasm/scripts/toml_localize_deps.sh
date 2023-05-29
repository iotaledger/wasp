#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
cd $contracts_path

# this doesn't work on Macbook
old_wasmlib="wasmlib = { git = \"https://github.com/iotaledger/wasp\", branch = \"develop\" }"
normal_new_wasmlib="wasmlib = { path = \"../../../../../packages/wasmvm/wasmlib\" }"
find . -mindepth 4 -maxdepth 4 -type f -name '*.toml' -exec sed -i "s|${old_wasmlib}|${normal_new_wasmlib}|g" {} +

gascalibration_new_wasmlib="wasmlib = { path = \"../../../../../../packages/wasmvm/wasmlib\" }"
find . -mindepth 5 -maxdepth 5 -type f -name '*.toml' -exec sed -i "s|${old_wasmlib}|${gascalibration_new_wasmlib}|g" {} +

old_wasmhost="wasmvmhost = { git = \"https://github.com/iotaledger/wasp\", branch = \"develop\" }"
normal_new_wasmhost="wasmvmhost = { path = \"../../../../../packages/wasmvm/wasmvmhost\" }"
find . -mindepth 4 -maxdepth 4 -type f -name '*.toml' -exec sed -i "s|${old_wasmhost}|${normal_new_wasmhost}|g" {} +

gascalibration_new_wasmhost="wasmvmhost = { path = \"../../../../../../packages/wasmvm/wasmvmhost\" }"
find . -mindepth 5 -maxdepth 5 -type f -name '*.toml' -exec sed -i "s|${old_wasmhost}|${gascalibration_new_wasmhost}|g" {} +
