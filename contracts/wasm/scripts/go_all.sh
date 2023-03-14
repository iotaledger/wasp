#!/bin/bash
contracts_path=$(git rev-parse --show-toplevel)/contracts/wasm
cd $contracts_path
go install ../../tools/schema
schema -go -build
