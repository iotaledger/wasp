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

wasp-client-owner fr admin init
fraddress=$(cat owner.json | jq .fr.address -r)
wasp-client-owner wallet send-funds $fraddress IOTA 100 # operating capital

wasp-client-owner fa admin init
faaddress=$(cat owner.json | jq .fa.address -r)
wasp-client-owner wallet send-funds $faaddress IOTA 100 # operating capital

wasp-client-owner tr admin init
traddress=$(cat owner.json | jq .tr.address -r)
wasp-client-owner wallet send-funds $traddress IOTA 100 # operating capital

wasp-client wallet init
wasp-client wallet request-funds
wasp-client fr set address $fraddress
wasp-client fa set address $faaddress
wasp-client tr set address $traddress

r=$(wasp-client tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

wasp-client fa start-auction "My first auction" "$color" 10 100 10
