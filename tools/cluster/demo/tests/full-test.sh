#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json wallet init
wwallet -c owner.json wallet request-funds

wwallet -c owner.json fr admin init
fraddress=$(cat owner.json | jq .fr.address -r)
wwallet -c owner.json wallet send-funds $fraddress IOTA 100 # operating capital

wwallet -c owner.json fa admin init
faaddress=$(cat owner.json | jq .fa.address -r)
wwallet -c owner.json wallet send-funds $faaddress IOTA 100 # operating capital

wwallet -c owner.json tr admin init
traddress=$(cat owner.json | jq .tr.address -r)
wwallet -c owner.json wallet send-funds $traddress IOTA 100 # operating capital

wwallet -c owner.json dwf admin init
dwfaddress=$(cat owner.json | jq .dwf.address -r)
wwallet -c owner.json wallet send-funds $dwfaddress IOTA 100 # operating capital

wwallet wallet init
wwallet wallet request-funds
wwallet fr set address $fraddress
wwallet fa set address $faaddress
wwallet tr set address $traddress
wwallet dwf set address $dwfaddress

r=$(wwallet tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

wwallet fa start-auction "My first auction" "$color" 10 100 10
wwallet fr bet 2 100
wwallet dwf donate 10 "cool app :)"
