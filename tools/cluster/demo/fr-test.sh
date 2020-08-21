#!/usr/bin/env bash
set -euo pipefail

ARGS="$*"

function wasp-client() {
    echo "wasp-client -w $ARGS $@"
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

echo "Waiting for set-period request to be executed..."
while true; do
    r=$(wasp-client-owner fr status | grep "play period") || true
    [[ "$r" =~ ": 10"$ ]] && break
    sleep 1
done

wasp-client wallet init
wasp-client wallet request-funds
wasp-client fr set address $scaddress

wasp-client fr bet 2 100

echo "Waiting for bet request to be executed..."
while true; do
    r=$(wasp-client fr status | grep "bets for next play") || true
    [[ "$r" =~ ": 1"$ ]] && break
    sleep 1
done
