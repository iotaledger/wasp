#!/bin/bash
example_name=$1 # path relative to wasp/contracts/wasm
flag=$2
node_modules_path=$(git rev-parse --show-toplevel)/contracts/wasm/node_modules
cd $example_name

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
