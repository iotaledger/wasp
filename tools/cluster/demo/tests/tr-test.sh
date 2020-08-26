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
wasp-client-owner tr admin init
scaddress=$(cat owner.json | jq .tr.address -r)
wasp-client-owner wallet send-funds $scaddress IOTA 100 # operating capital

wasp-client wallet init
wasp-client wallet request-funds
wasp-client tr set address $scaddress

r=$(wasp-client tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

# verify
wasp-client tr status | tee >(cat >&2) | grep -q 'Supply: 10$'
