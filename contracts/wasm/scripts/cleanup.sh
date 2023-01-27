#!/bin/bash
example_name=$1
cd $example_name
if [[ ! -f "schema.yaml" ]]; then
    echo 'schema.yaml not exists'
    exit
fi
schema -go -rs -ts -clean
rm ./ts/"$example_name"impl/tsconfig.json
rm ./rs/"$example_name"impl/Cargo.lock
rm ./rs/"$example_name"impl/Cargo.toml
rm ./rs/"$example_name"impl/LICENSE
rm ./rs/"$example_name"impl/README.md
cd ..
