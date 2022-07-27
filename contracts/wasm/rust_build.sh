#!/bin/bash
example_name=$1
flag=$2
cd $example_name

if [ ! -f "schema.yaml" ]; then
    exit 1
fi

echo "Building $example_name"
schema -rust $flag
echo "compiling "$example_name"_bg.wasm"
wasm-pack build
