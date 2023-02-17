#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

cd contracts/wasm/scripts
./cleanup_all.sh
./all_build.sh
./update_hardcoded.sh

cd ${CURRENT_DIR}
