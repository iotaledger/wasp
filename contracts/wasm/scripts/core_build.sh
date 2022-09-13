#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
cd $root_path/contracts/wasm
npm install

go install $root_path/tools/schema
cd $root_path/packages/wasmvm/wasmlib
schema -core -go -rust -ts -force
cd $root_path/contracts/wasm
rm -rf ./node_modules/wasmlib/
cp -R $root_path/packages/wasmvm/wasmlib/ts/wasmlib ./node_modules
