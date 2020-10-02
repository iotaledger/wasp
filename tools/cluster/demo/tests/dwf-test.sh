#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json init
wwallet -c owner.json request-funds
wwallet -c owner.json dwf admin deploy
scaddress=$(cat owner.json | jq .sc.dwf.address -r)
wwallet -c owner.json send-funds $scaddress IOTA 100 # operating capital

wwallet init
wwallet request-funds
wwallet dwf set address $scaddress

wwallet dwf donate 10 "donation 1"
wwallet dwf donate 100 "donation 2"

# check that donate request has been executed
wwallet dwf status | tee >(cat >&2) | grep -q 'donation 1$'
wwallet dwf status | tee >(cat >&2) | grep -q 'donation 2$'

wwallet -c owner.json dwf withdraw 20
