#!/bin/bash
cd ..
go install ../../tools/schema
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash scripts/schema_build.sh "$dir" $1
  fi
done
cd gascalibration
for dir in ./*; do
  if [ -d "$dir" ]; then
    bash ../scripts/schema_build.sh "$dir" $1
  fi
done
cd ../scripts
