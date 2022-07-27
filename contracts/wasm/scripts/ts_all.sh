#!/bin/bash
node_modules_path="../node_modules"
cd ..
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash scripts/ts_build.sh "$dir" "$node_modules_path" $1
  fi
done
cd gascalibration
node_modules_path="../../node_modules"
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash ../scripts/ts_build.sh "$dir" "$node_modules_path" $1
  fi
done
cd ../scripts
