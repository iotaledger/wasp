#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm

cd $contracts_path
npm install

go install $root_path/tools/schema

cd $root_path/packages/wasmvm/wasmlib
schema -go -rs -ts -force

cd $contracts_path
rm -rf ./node_modules/wasmlib/
cp -R $root_path/packages/wasmvm/wasmlib/as/wasmlib ./node_modules
rm -rf ./node_modules/wasmvmhost/
cp -R $root_path/packages/wasmvm/wasmvmhost/ts/wasmvmhost ./node_modules
