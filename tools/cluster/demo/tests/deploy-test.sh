#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

wwallet -c owner.json init
wwallet -c owner.json request-funds

wwallet --sc=tr -c owner.json sc deploy '0,1,2,3' 3 '8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p' 'TokenRegistry'
scaddress=$(cat owner.json | jq .sc.tr.address -r)

wwallet -c owner.json send-funds $scaddress IOTA 100 # operating capital

wwallet init
wwallet request-funds
wwallet tr set address $scaddress

r=$(wwallet tr mint "My first coin" 10)
echo "$r"
[[ "$r" =~ of[[:space:]]color[[:space:]]([[:alnum:]]+) ]]
color=${BASH_REMATCH[1]}

# verify
wwallet tr query "$color" | tee >(cat >&2) | grep -q 'Supply: 10$'
