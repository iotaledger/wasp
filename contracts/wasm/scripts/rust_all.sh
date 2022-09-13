#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
# build all the schema.yaml, since any missing members of rust would cause contract building fail
bash $contracts_path/scripts/schema_all.sh

cd $contracts_path
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash scripts/rust_build.sh "$dir" $1
  fi
done
cd $contracts_path/gascalibration
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash ../scripts/rust_build.sh "$dir" $1
  fi
done
