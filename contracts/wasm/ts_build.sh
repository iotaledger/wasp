#!/bin/bash
example_name=$1
flag=$2
cd $example_name

if [ -f "schema.yaml" ]; then
    if [ -f "schema.json" ]; then
        exit 1
    fi
fi

echo "Building $example_name"
schema -ts $flag
echo "compiling "$example_name"_ts.wasm"
npx asc ts/"$example_name"/lib.ts --lib ../node_modules -O --outFile ts/pkg/"$example_name"_ts.wasm
