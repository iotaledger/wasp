#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json wallet init
wwallet -c owner.json wallet request-funds
wwallet -c owner.json tr admin init
scaddress=$(cat owner.json | jq .tr.address -r)
wwallet -c owner.json wallet send-funds $scaddress IOTA 100 # operating capital

wwallet wallet init
wwallet wallet request-funds
wwallet tr set address $scaddress

r=$(wwallet tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

# verify
wwallet tr query "$color" | tee >(cat >&2) | grep -q 'Supply: 10$'
