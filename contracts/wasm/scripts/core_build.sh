#!/bin/bash
go install ../../../tools/schema
cd ../../../packages/wasmvm/wasmlib
schema -core -go -rust -ts -force
cd ../../../contracts/wasm
rm -rf ./node_modules/wasmlib/
cp -R ../../packages/wasmvm/wasmlib/ts/wasmlib ./node_modules
cd scripts
