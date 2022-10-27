#!/bin/bash
example_name=$1 # path relative to wasp/contracts/wasm
flag=$2
cd $example_name

if [ ! -f "schema.yaml" ]; then
  echo "schema.yaml not found"
  exit 1
fi

echo "Building $example_name"
schema -rust $flag
echo "Compiling "$example_name"_main_bg.wasm"
wasm-pack build "rs/"$example_name"_main"

