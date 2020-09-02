#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json wallet init
wwallet -c owner.json wallet request-funds
wwallet -c owner.json fa admin init
scaddress=$(cat owner.json | jq .fa.address -r)
wwallet -c owner.json wallet send-funds $scaddress IOTA 100 # operating capital

wwallet wallet init
wwallet wallet request-funds
wwallet fa set address $scaddress

r=$(wwallet wallet mint 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

wwallet fa start-auction "My first auction" "$color" 10 100 10

# check that start-auction request has been executed
wwallet fa status | tee >(cat >&2) | grep -q -- "- color: $color\$"

wwallet fa place-bid "$color" 110

# check that place-bid request has been executed
wwallet fa status | tee >(cat >&2) | grep -q "bidder:"
