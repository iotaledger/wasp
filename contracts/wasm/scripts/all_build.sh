#!/bin/bash
root_path=$(git rev-parse --show-toplevel)/contracts/wasm/scripts
bash $root_path/core_build.sh
bash $root_path/schema_all.sh
bash $root_path/go_all.sh -force
bash $root_path/ts_all.sh -force
bash $root_path/rust_all.sh -force
bash $root_path/update_hardcoded.sh
