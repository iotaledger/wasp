#!/bin/bash
if [ -f "../testcore/pkg/testcore_bg.wasm" ]; then
    cp ../testcore/pkg/testcore_bg.wasm ../../../packages/vm/core/testcore/sbtests/sbtestsc/
fi
if [ -f "../inccounter/pkg/inccounter_bg.wasm" ]; then
    cp ../inccounter/pkg/inccounter_bg.wasm ../../../tools/cluster/tests/wasm/
fi

cd ../../../documentation/tutorial-examples
wasm-pack build
cp pkg/solotutorial_bg.wasm test/
cd ../../contracts/wasm/scripts
