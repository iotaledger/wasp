#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm
if [ -f "$contracts_path/testcore/rs/main/pkg/main_bg.wasm" ]; then
    cp $contracts_path/testcore/rs/main/pkg/main_bg.wasm $root_path/packages/vm/core/testcore/sbtests/sbtestsc/testcore_bg.wasm
fi
if [ -f "$contracts_path/inccounter/rs/main/pkg/main_bg.wasm" ]; then
    cp $contracts_path/inccounter/rs/main/pkg/main_bg.wasm $root_path/tools/cluster/tests/wasm/inccounter_bg.wasm
fi

cd $root_path/documentation/tutorial-examples
wasm-pack build
cp pkg/solotutorial_bg.wasm test/
