#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet init
wwallet request-funds
wwallet chain deploy --chain=chain1 --committee='0,1,2,3' --quorum=3

vmtype=wasmtimevm
name=inccounter
description="inccounter SC"
file="$DIR/../../tests/wasptest_new/wasm/inccounter_bg.wasm"

wwallet chain deploy-contract "$vmtype" "$name" "$description" "$file"

# check that new contract is listed
r=$(wwallet chain list-contracts)
echo "$r"
[[ $(echo "$r" | wc -l) == "4" ]]

r=$(wwallet chain call-view "$name" incrementViewCounter | wwallet decode string counter int)
[[ "$r" == "counter: 0" ]]

wwallet chain post-request "$name" increment

r=$(wwallet chain call-view "$name" incrementViewCounter | wwallet decode string counter int)
[[ "$r" == "counter: 1" ]]

echo "PASS"
