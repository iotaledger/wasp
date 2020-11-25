#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

alias="chain1"

wwallet init
wwallet request-funds

[[ $(wwallet chain list | wc -l) == "0" ]]

wwallet chain deploy --chain=$alias --committee='0,1,2,3' --quorum=3
chainid=$(cat wwallet.json | jq .chains.$alias -r)

[[ $(wwallet chain list | wc -l) == "1" ]]
[[ $(wwallet chain list) =~ "$chainid" ]]

# unnecessary, since it is the latest deployed chain
wwallet set chain $alias

r=$(wwallet chain info)
echo "$r"
[[ "$r" =~ "$chainid" ]]

echo "PASS"
