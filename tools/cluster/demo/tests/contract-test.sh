#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli init
wasp-cli request-funds
wasp-cli chain deploy --chain=chain1 --committee='0,1,2,3' --quorum=3 --description="Test chain"

vmtype=wasmtimevm
name=inccounter
description="inccounter SC"
file="$DIR/../../tests/wasptest_new/wasm/inccounter_bg.wasm"

wasp-cli chain deploy-contract "$vmtype" "$name" "$description" "$file"

# check that new contract is listed
r=$(wasp-cli chain list-contracts)
echo "$r"
[[ $(echo "$r" | tail -n +5 | wc -l) == "4" ]]

r=$(wasp-cli chain call-view "$name" increment_view_counter | wasp-cli decode string counter int)
[[ "$r" == "counter: 0" ]]

wasp-cli chain post-request "$name" increment

r=$(wasp-cli chain call-view "$name" increment_view_counter | wasp-cli decode string counter int)
[[ "$r" == "counter: 1" ]]

echo "PASS"
