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

wwallet-owner fr admin init
fraddress=$(cat owner.json | jq .fr.address -r)
wwallet-owner wallet send-funds $fraddress IOTA 100 # operating capital

wwallet-owner fa admin init
faaddress=$(cat owner.json | jq .fa.address -r)
wwallet-owner wallet send-funds $faaddress IOTA 100 # operating capital

wwallet-owner tr admin init
traddress=$(cat owner.json | jq .tr.address -r)
wwallet-owner wallet send-funds $traddress IOTA 100 # operating capital

wwallet wallet init
wwallet wallet request-funds
wwallet fr set address $fraddress
wwallet fa set address $faaddress
wwallet tr set address $traddress

r=$(wwallet tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]](.+)$ ]]
color=${BASH_REMATCH[1]}

wwallet fa start-auction "My first auction" "$color" 10 100 10
