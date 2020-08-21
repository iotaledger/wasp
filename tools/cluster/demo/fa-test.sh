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

echo "Waiting for start-auction request to be executed..."
while true; do
    r=$(wasp-client fa status | grep "color:") || true
    [ "$r" ] && break
    sleep 1
done

wasp-client fa place-bid "$color" 110

echo "Waiting for start-auction request to be executed..."
while true; do
    r=$(wasp-client fa status | grep "bidder:") || true
    [ "$r" ] && break
    sleep 1
done

