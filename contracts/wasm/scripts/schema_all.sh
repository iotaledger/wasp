#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
go install $contracts_path/tools/schema
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
