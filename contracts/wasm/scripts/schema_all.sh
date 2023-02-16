#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm

go install $root_path/tools/schema

cd $contracts_path
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash scripts/schema_build.sh "$dir" $1
  fi
done
cd $contracts_path/gascalibration
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash ../scripts/schema_build.sh "$dir" $1
  fi
done
