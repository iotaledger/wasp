#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm
if [ -f "$contracts_path/testcore/pkg/testcore_bg.wasm" ]; then
    cp $contracts_path/testcore/pkg/testcore_bg.wasm $root_path/packages/vm/core/testcore/sbtests/sbtestsc/
fi
if [ -f "$contracts_path/inccounter/pkg/inccounter_bg.wasm" ]; then
    cp $contracts_path/inccounter/pkg/inccounter_bg.wasm $root_path/tools/cluster/tests/wasm/
fi

cd $root_path/documentation/tutorial-examples
wasm-pack build
cp pkg/solotutorial_bg.wasm test/
