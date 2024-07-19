#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
cd $contracts_path
go install ../../tools/schema
schema -rs
if [ "$1" == "ci" ]; then
    # in CI all, using local dependency will make things easier
    bash ./scripts/toml_localize_deps.sh
fi
schema -rs -build
