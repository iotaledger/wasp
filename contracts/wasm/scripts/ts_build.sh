#!/bin/bash
example_path=$1
flag=$2
cd $example_path
example_name=$(basename $example_path) # it is path relative to wasp/contracts/wasm in the meantime
node_modules_path=$(git rev-parse --show-toplevel)/contracts/wasm/node_modules

if [ ! -f "schema.yaml" ]; then
  echo "schema.yaml not found"
  exit 1
fi

echo "Building $example_name"
schema -ts $flag
echo "Compiling "$example_name"_ts.wasm"
if [ ! -d "./ts/pkg" ]; then
  mkdir ./ts/pkg
fi
npx asc ts/main.ts --lib "$node_modules_path" -O --outFile ts/pkg/"$example_name"_ts.wasm
