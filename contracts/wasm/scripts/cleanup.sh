#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
cd $contracts_path
go install ../../tools/schema
find . -name "Cargo.lock" -type f -delete
schema -go -rs -ts -clean
