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

# test the list-blobs command
r=$(wasp-cli chain list-blobs)
echo "$r"
[[ $(echo "$r" | head -n 1) =~ "Total 0 blob(s)" ]]

# test the store-blob command
wasp-cli chain store-blob string p file "$file" string v string "$vmtype" string d string "$description"

# test the list-blobs command
r=$(wasp-cli chain list-blobs)
echo "$r"
[[ $(echo "$r" | head -n 1) =~ "Total 1 blob(s)" ]]

blobhash=$(echo "$r" | tail -n +5 | sed 's/[:[:space:]].*$//')

# test the show-blob command
r=$(wasp-cli chain show-blob "$blobhash" | wasp-cli decode string d string)
echo "$r"
echo "$r" | grep -q "$description"

echo "PASS"
