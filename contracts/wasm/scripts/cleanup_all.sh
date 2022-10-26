#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
cd $contracts_path
for dir in ./*; do
  if [ -d "$dir" ]; then
    echo "$dir"
    bash scripts/cleanup.sh "$dir"
  fi
done
cd $contracts_path/gascalibration
for dir in ./*; do
  if [ -d "$dir" ]; then
    echo "$dir"
    bash ../scripts/cleanup.sh "$dir"
  fi
done
