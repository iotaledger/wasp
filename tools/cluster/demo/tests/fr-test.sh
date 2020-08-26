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
wasp-client-owner fr admin set-period 10
scaddress=$(cat owner.json | jq .fr.address -r)
wasp-client-owner wallet send-funds $scaddress IOTA 100 # operating capital

# check that set-period request has been executed
wasp-client-owner fr status | tee >(cat >&2) | grep -q 'play period.*: 10$'

wasp-client wallet init
wasp-client wallet request-funds
wasp-client fr set address $scaddress

wasp-client fr bet 2 100

# check that bet request has been executed
wasp-client fr status | tee >(cat >&2) | grep -q 'bets for next play: 1$'
