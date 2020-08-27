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
wwallet-owner fa admin init
scaddress=$(cat owner.json | jq .fa.address -r)
wwallet-owner wallet send-funds $scaddress IOTA 100 # operating capital

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
