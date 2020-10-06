#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json init
wwallet -c owner.json request-funds
wwallet -c owner.json fa admin deploy
scaddress=$(cat owner.json | jq .sc.fa.address -r)
wwallet -c owner.json send-funds $scaddress IOTA 100 # operating capital

wwallet init
wwallet request-funds
wwallet fa set address $scaddress

r=$(wwallet mint 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]]([[:alnum:]]+)$ ]]
color=${BASH_REMATCH[1]}

wwallet fa start-auction "My first auction" "$color" 10 100 10

# check that start-auction request has been executed
wwallet fa status | tee >(cat >&2) | grep -q -- "- color: $color\$"

wwallet fa place-bid "$color" 110

# check that place-bid request has been executed
wwallet fa status | tee >(cat >&2) | grep -q "bidder:"
