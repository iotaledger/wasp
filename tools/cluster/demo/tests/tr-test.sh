#!/usr/bin/env bash
set -euo pipefail

err_report() {
    echo "Error on line $1" >&2
}

trap 'err_report $LINENO' ERR

ARGS="$*"

function wwallet() {
    echo "wwallet -w $ARGS $@" >&2
    command wwallet -w $ARGS "$@"
}

function wwallet-owner() {
    wwallet -c owner.json "$@"
}

wwallet-owner wallet init
wwallet-owner wallet request-funds
wwallet-owner tr admin init
scaddress=$(cat owner.json | jq .tr.address -r)
wwallet-owner wallet send-funds $scaddress IOTA 100 # operating capital

wwallet wallet init
wwallet wallet request-funds
wwallet tr set address $scaddress

r=$(wwallet tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

# verify
wwallet tr query "$color" | tee >(cat >&2) | grep -q 'Supply: 10$'
