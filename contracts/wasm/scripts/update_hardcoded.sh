#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm
if [ -f "$contracts_path/testcore/rs/testcore_main/pkg/testcore_main_bg.wasm" ]; then
    cp $contracts_path/testcore/rs/testcore_main/pkg/testcore_main_bg.wasm $root_path/packages/vm/core/testcore/sbtests/sbtestsc/testcore_bg.wasm
fi
if [ -f "$contracts_path/inccounter/rs/inccounter_main/pkg/inccounter_main_bg.wasm" ]; then
    cp $contracts_path/inccounter/rs/inccounter_main/pkg/inccounter_main_bg.wasm $root_path/tools/cluster/tests/wasm/inccounter_bg.wasm
fi

cd $root_path/documentation/tutorial-examples
wasm-pack build rs/solotutorial_main
cp rs/solotutorial_main/pkg/solotutorial_main_bg.wasm test/solotutorial_bg.wasm
