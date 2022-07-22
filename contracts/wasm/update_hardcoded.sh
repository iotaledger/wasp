#!/bin/bash
for dir in ./*; do
    if [ -d "$dir/pkg/"$dir"_bg.wasm" ]; then
        cp $dir/pkg/$dir_bg.wasm $dir/test/
    fi
done
cd gascalibration
for dir in ./*; do
    if [ -d "$dir/pkg/"$dir"_bg.wasm" ]; then
        cp $dir/pkg/$dir_bg.wasm $dir/test/
    fi
    if [ -d "$dir/go/pkg/"$dir"_go.wasm" ]; then
        cp $dir/go/pkg/$dir_go.wasm $dir/test/
    fi
    if [ -d "$dir/ts/pkg/"$dir"_ts.wasm" ]; then
        cp $dir/ts/pkg/$dir_ts.wasm $dir/test/
    fi
done
cd ..
if [ -f "testcore/pkg/testcore_bg.wasm" ]; then
    cp testcore/pkg/testcore_bg.wasm ../../packages/vm/core/testcore/sbtests/sbtestsc/
fi
if [ -f "inccounter/pkg/inccounter_bg.wasm" ]; then
    cp inccounter/pkg/inccounter_bg.wasm ../../tools/cluster/tests/wasm/
fi

cd ../../documentation/tutorial-examples
wasm-pack build
cp pkg/solotutorial_bg.wasm test/
cd ../../contracts/wasm/
