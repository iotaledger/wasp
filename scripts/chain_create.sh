#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

echo -e "\nrequesting funds..."
./wasp-cli request-funds
echo -e "\ndeploying chain..."
./wasp-cli chain deploy --chain=testchain --quorum=1 --committee=0 --verbose
echo -e "\ndepositing to chain..."
./wasp-cli chain deposit 0x8B65DD08C7784017fe6B8Af20904e61916506fD4 base:100000 -w=false

cd ${CURRENT_DIR}
