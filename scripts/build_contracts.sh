#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

cd tools/schema
go install
cd ${PARENT_DIR}

make install
cd contracts/wasm/scripts
./cleanup.sh
./all_build.sh
./update_hardcoded.sh

cd ${CURRENT_DIR}
