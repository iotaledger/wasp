#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
root_path=$(git rev-parse --show-toplevel)/contracts/wasm/scripts
bash $root_path/core_build.sh
cd $contracts_path
find . -name "Cargo.lock" -type f -delete
schema -go -rs -ts -force
schema -go -rs -ts -build
bash $root_path/update_hardcoded.sh
