#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json init
wwallet -c owner.json request-funds
wwallet -c owner.json fr admin init
wwallet -c owner.json fr admin set-period 10
scaddress=$(cat owner.json | jq .fr.address -r)
wwallet -c owner.json send-funds $scaddress IOTA 100 # operating capital

# check that set-period request has been executed
wwallet -c owner.json fr status | tee >(cat >&2) | grep -q 'play period.*: 10$'

wwallet init
wwallet request-funds
wwallet fr set address $scaddress

wwallet fr bet 2 100

# check that bet request has been executed
wwallet fr status | tee >(cat >&2) | grep -q 'bets for next play: 1$'
