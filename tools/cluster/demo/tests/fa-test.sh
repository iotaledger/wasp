#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli -c owner.json init
wasp-cli -c owner.json request-funds
wasp-cli -c owner.json fa admin deploy
scaddress=$(cat owner.json | jq .sc.fa.address -r)
wasp-cli -c owner.json send-funds $scaddress IOTA 100 # operating capital

wasp-cli init
wasp-cli request-funds
wasp-cli fa set address $scaddress

r=$(wasp-cli mint 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]]([[:alnum:]]+)$ ]]
color=${BASH_REMATCH[1]}

wasp-cli fa start-auction "My first auction" "$color" 10 100 10

# check that start-auction request has been executed
wasp-cli fa status | tee >(cat >&2) | grep -q -- "- color: $color\$"

wasp-cli fa place-bid "$color" 110

# check that place-bid request has been executed
wasp-cli fa status | tee >(cat >&2) | grep -q "bidder:"
