#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet init
wwallet request-funds
wwallet chain deploy --chain=chain1 --committee='0,1,2,3' --quorum=3

vmtype=wasmtimevm
name=increment
description="increment SC"
file="$DIR/../../tests/wasptest_new/wasm/increment_bg.wasm"

wwallet chain deploy-contract "$vmtype" "$name" "$description" "$file"

# check that new contract is listed
r=$(wwallet chain list-contracts)
echo "$r"
[[ $(echo "$r" | wc -l) == "4" ]]

echo "PASS"
