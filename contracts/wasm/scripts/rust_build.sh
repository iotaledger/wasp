#!/bin/bash
example_path=$1
flag=$2
cd $example_path
example_name=$(basename $example_path) # it is path relative to wasp/contracts/wasm in the meantime

echo $example_name
if [ ! -f "schema.yaml" ]; then
  echo "schema.yaml not found"
  exit 1
fi

echo "Building $example_name"
schema -rs $flag
echo "Compiling "$example_name"wasm_bg.wasm"
wasm-pack build "rs/"$example_name"wasm"

