#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli -c owner.json init
wasp-cli -c owner.json request-funds

wasp-cli -c owner.json fr admin deploy
fraddress=$(cat owner.json | jq .sc.fr.address -r)
wasp-cli -c owner.json send-funds $fraddress IOTA 100 # operating capital

wasp-cli -c owner.json fa admin deploy
faaddress=$(cat owner.json | jq .sc.fa.address -r)
wasp-cli -c owner.json send-funds $faaddress IOTA 100 # operating capital

wasp-cli -c owner.json tr admin deploy
traddress=$(cat owner.json | jq .sc.tr.address -r)
wasp-cli -c owner.json send-funds $traddress IOTA 100 # operating capital

wasp-cli -c owner.json dwf admin deploy
dwfaddress=$(cat owner.json | jq .sc.dwf.address -r)
wasp-cli -c owner.json send-funds $dwfaddress IOTA 100 # operating capital

wasp-cli init
wasp-cli request-funds
wasp-cli fr set address $fraddress
wasp-cli fa set address $faaddress
wasp-cli tr set address $traddress
wasp-cli dwf set address $dwfaddress

r=$(wasp-cli tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]]([[:alnum:]]+) ]]
color=${BASH_REMATCH[1]}

wasp-cli fa start-auction "My first auction" "$color" 10 100 10
wasp-cli fr bet 2 100
wasp-cli dwf donate 10 "cool app :)"
