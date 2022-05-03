#!/bin/bash
for dir in ./*; do
    if [ -d "$dir/pkg/"$dir"_bg.wasm" ]; then
        cp $dir/pkg/$dir_bg.wasm $dir/test/
    fi
done
if [ -d "testcore/pkg/testcore_bg.wasm" ]; then
    cp testcore/pkg/testcore_bg.wasm ../../packages/vm/core/testcore_stardust/sbtests/sbtestsc/
fi
if [ -d "inccounter/pkg/inccounter_bg.wasm" ]; then
    cp inccounter/pkg/inccounter_bg.wasm ../../tools/cluster/tests/wasm/
fi

cd ../../documentation/tutorial-examples
wasm-pack build
cp pkg/example_tutorial_bg.wasm test/
cd ../../contracts/wasm/
