#!/bin/bash
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash schema_build.sh "$dir" $1
  fi
done
cd gascalibration
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash ../schema_build.sh "$dir" $1
  fi
done
cd ..
