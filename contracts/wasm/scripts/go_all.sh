#!/bin/bash
cd ..
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash scripts/go_build.sh "$dir" $1
  fi
done
cd gascalibration
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash ../scripts/go_build.sh "$dir" $1
  fi
done
cd ../scripts
