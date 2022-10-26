#!/bin/bash
example_name=$1
cd $example_name
schema -go -rust -ts -clean
rm ./ts/$example_name/tsconfig.json
cd ..
