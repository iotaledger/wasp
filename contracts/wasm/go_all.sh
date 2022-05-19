#!/bin/bash
for dir in ./*; do
 if [ -d "$dir" ]; then
    bash go_build.sh "$dir" $1
  fi
done
