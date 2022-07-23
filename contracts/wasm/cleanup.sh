#!/bin/bash
root=$(git rev-parse --show-toplevel)

 rm "$root"/contracts/wasm/*/*/*/consts.*
 rm "$root"/contracts/wasm/*/*/*/contract.*
 rm "$root"/contracts/wasm/*/*/*/keys.*
 rm "$root"/contracts/wasm/*/*/*/lib.*
 rm "$root"/contracts/wasm/*/*/*/params.*
 rm "$root"/contracts/wasm/*/*/*/results.*
 rm "$root"/contracts/wasm/*/*/*/state.*
 rm "$root"/contracts/wasm/*/*/*/typedefs.*
 rm "$root"/contracts/wasm/*/*/*/types.*
 rm -r "$root"/contracts/wasm/*/*/pkg
 rm -r "$root"/contracts/wasm/target
 rm -r "$root"/contracts/wasm/node_modules

for dir in "$root"/contracts/wasm/*; do
 if [ -d "$dir" ]; then
    rm "$dir"/go/main.go
    rm "$dir"/ts/"$dir"/index.ts
    rm "$dir"/ts/"$dir"/tsconfig.json
    rm "$dir"/pkg/*.*
    rm "$dir"/ts/pkg/*.*
  fi
done

cd gascalibration
for dir in ./*; do
 if [ -d "$dir" ]; then
   rm "$dir"/go/main.go
    rm "$dir"/ts/"$dir"/index.ts
    rm "$dir"/ts/"$dir"/tsconfig.json
    rm "$dir"/pkg/*.*
    rm "$dir"/ts/pkg/*.*
  fi
done
cd ..
