#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
root_path=$(git rev-parse --show-toplevel)/contracts/wasm/scripts
bash $root_path/core_build.sh
cd $contracts_path
schema -go -rs -ts -force
schema -go -rs -ts -build
cd ../..
bash $root_path/update_hardcoded.sh
