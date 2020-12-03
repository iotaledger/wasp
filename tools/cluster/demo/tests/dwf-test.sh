#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli -c owner.json init
wasp-cli -c owner.json request-funds
wasp-cli -c owner.json dwf admin deploy
scaddress=$(cat owner.json | jq .sc.dwf.address -r)
wasp-cli -c owner.json send-funds $scaddress IOTA 100 # operating capital

wasp-cli init
wasp-cli request-funds
wasp-cli dwf set address $scaddress

wasp-cli dwf donate 10 "donation 1"
wasp-cli dwf donate 100 "donation 2"

# check that donate request has been executed
wasp-cli dwf status | tee >(cat >&2) | grep -q 'donation 1$'
wasp-cli dwf status | tee >(cat >&2) | grep -q 'donation 2$'

wasp-cli -c owner.json dwf withdraw 20
