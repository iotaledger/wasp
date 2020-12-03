#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli -c owner.json init
wasp-cli -c owner.json request-funds
wasp-cli -c owner.json tr admin deploy
scaddress=$(cat owner.json | jq .sc.tr.address -r)
wasp-cli -c owner.json send-funds $scaddress IOTA 100 # operating capital

wasp-cli init
wasp-cli request-funds
wasp-cli tr set address $scaddress

r=$(wasp-cli tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]]([[:alnum:]]+) ]]
color=${BASH_REMATCH[1]}

# verify
wasp-cli tr query "$color" | tee >(cat >&2) | grep -q 'Supply: 10$'
