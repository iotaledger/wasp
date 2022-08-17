#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
node_modules_path="../node_modules"
cd $contracts_path
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash scripts/ts_build.sh "$dir" "$node_modules_path" $1
  fi
done
cd $contracts_path/gascalibration
node_modules_path="../../node_modules"
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash ../scripts/ts_build.sh "$dir" "$node_modules_path" $1
  fi
done
