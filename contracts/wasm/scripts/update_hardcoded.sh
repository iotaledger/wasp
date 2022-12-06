#!/bin/bash
root_path=$(git rev-parse --show-toplevel)
contracts_path=$root_path/contracts/wasm
if [ -f "$contracts_path/testcore/rs/testcorewasm/pkg/testcorewasm_bg.wasm" ]; then
    cp $contracts_path/testcore/rs/testcorewasm/pkg/testcorewasm_bg.wasm $root_path/packages/vm/core/testcore/sbtests/sbtestsc/testcore_bg.wasm
fi
if [ -f "$contracts_path/inccounter/rs/inccounterwasm/pkg/inccounterwasm_bg.wasm" ]; then
    cp $contracts_path/inccounter/rs/inccounterwasm/pkg/inccounterwasm_bg.wasm $root_path/tools/cluster/tests/wasm/inccounter_bg.wasm
fi

cd $root_path/documentation/tutorial-examples
wasm-pack build rs/solotutorialwasm
cp rs/solotutorialwasm/pkg/solotutorialwasm_bg.wasm test/solotutorial_bg.wasm
