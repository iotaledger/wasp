#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
cd $root_path/contracts/wasm
npm install

go install $root_path/tools/schema
cd $root_path/packages/wasmvm/wasmlib
schema -core -go -rs -ts -force
cd $root_path/contracts/wasm
rm -rf ./node_modules/wasmlib/
cp -R $root_path/packages/wasmvm/wasmlib/as/wasmlib ./node_modules
rm -rf ./node_modules/wasmvmhost/
cp -R $root_path/packages/wasmvm/wasmvmhost/ts/wasmvmhost ./node_modules
