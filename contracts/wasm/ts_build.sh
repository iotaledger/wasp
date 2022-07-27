#!/bin/bash
example_name=$1
node_modules_path=$2
flag=$3

cd $example_name

if [ ! -f "schema.yaml" ]; then
    echo "schema.yaml not found"
    exit 1
fi

echo "Building $example_name"
schema -ts $flag
echo "compiling "$example_name"_ts.wasm"
npx asc ts/"$example_name"/lib.ts --lib "$node_modules_path" -O --outFile ts/pkg/"$example_name"_ts.wasm
