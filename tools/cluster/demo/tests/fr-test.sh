#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wasp-cli -c owner.json init
wasp-cli -c owner.json request-funds
wasp-cli -c owner.json fr admin deploy
wasp-cli -c owner.json fr admin set-period 10
scaddress=$(cat owner.json | jq .sc.fr.address -r)
wasp-cli -c owner.json send-funds $scaddress IOTA 100 # operating capital

# check that set-period request has been executed
wasp-cli -c owner.json fr status | tee >(cat >&2) | grep -q 'play period.*: 10$'

wasp-cli init
wasp-cli request-funds
wasp-cli fr set address $scaddress

wasp-cli fr bet 2 100

# check that bet request has been executed
wasp-cli fr status | tee >(cat >&2) | grep -q 'bets for next play: 1$'
