#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
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
