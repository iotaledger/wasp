#!/bin/bash
example_name=$1
flag=$2
cd $example_name

if [ ! -f "schema.yaml" ]; then
  exit 1
fi

echo "Generating $example_name"
schema -go -rust -ts $flag
