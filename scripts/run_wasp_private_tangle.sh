#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

make build
./wasp -c config_defaults.json --webapi.auth.scheme=none --inx.address=localhost:9011

cd ${CURRENT_DIR}
