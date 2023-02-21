#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

OUTPUT_TO_FILE=false
if [[ $1 = "-f" ]]; then
    OUTPUT_TO_FILE=true
fi    

# check if richgo can be used to print colorful outputs
GO_EXECUTABLE=go
if [ "$OUTPUT_TO_FILE" = false ] && [ -x "$(command -v richgo)" ]; then
    GO_EXECUTABLE=richgo
fi

FILES=$(go list ./... | grep -v github.com/iotaledger/wasp/contracts/wasm)
${GO_EXECUTABLE} clean -testcache

make wasm
make install

if [ "$OUTPUT_TO_FILE" = false ]; then
    ${GO_EXECUTABLE} test -p 1 -failfast -count=1 -timeout=90m -short ${FILES}
else
    ${GO_EXECUTABLE} test -p 1 -failfast -count=1 -timeout=90m -short -v ${FILES} 2>&1 | tee tests_output.log
fi

cd ${CURRENT_DIR}
