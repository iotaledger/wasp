#!/usr/bin/env bash
set -euo pipefail

err_report() {
    echo "Error on line $1" >&2
}

trap 'err_report $LINENO' ERR

ARGS="$*"

function wasp-client() {
    echo "wasp-client -w $ARGS $@" >&2
    command wasp-client -w $ARGS "$@"
}

function wasp-client-owner() {
    wasp-client -c owner.json "$@"
}

wasp-client-owner wallet init
wasp-client-owner wallet request-funds
wasp-client-owner fa admin init
scaddress=$(cat owner.json | jq .fa.address -r)
wasp-client-owner wallet send-funds $scaddress IOTA 100 # operating capital

wasp-client wallet init
wasp-client wallet request-funds
wasp-client fa set address $scaddress

r=$(wasp-client wallet mint 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

wasp-client fa start-auction "My first auction" "$color" 10 100 10

# check that start-auction request has been executed
wasp-client fa status | tee >(cat >&2) | grep -q -- "- color: $color\$"

wasp-client fa place-bid "$color" 110

# check that place-bid request has been executed
wasp-client fa status | tee >(cat >&2) | grep -q "bidder:"
