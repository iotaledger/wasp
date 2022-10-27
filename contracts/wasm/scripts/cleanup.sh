#!/bin/bash
example_name=$1
cd $example_name
schema -go -rust -ts -clean
rm ./ts/$example_name/tsconfig.json
rm ./rs/$example_name/Cargo.lock
rm ./rs/$example_name/Cargo.toml
rm ./rs/$example_name/LICENSE
rm ./rs/$example_name/README.md
cd ..
