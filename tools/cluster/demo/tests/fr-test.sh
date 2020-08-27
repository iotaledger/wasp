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
wwallet-owner fr admin set-period 10
scaddress=$(cat owner.json | jq .fr.address -r)
wwallet-owner wallet send-funds $scaddress IOTA 100 # operating capital

# check that set-period request has been executed
wwallet-owner fr status | tee >(cat >&2) | grep -q 'play period.*: 10$'

wwallet wallet init
wwallet wallet request-funds
wwallet fr set address $scaddress

wwallet fr bet 2 100

# check that bet request has been executed
wwallet fr status | tee >(cat >&2) | grep -q 'bets for next play: 1$'
