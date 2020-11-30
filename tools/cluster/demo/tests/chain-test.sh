#!/usr/bin/env bash
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/common.sh"

alias="chain1"

wwallet init
wwallet request-funds

r=$(wwallet address)
echo "$r"
[[ "$r" =~ Address:[[:space:]]+([[:alnum:]]+)$ ]]
owneraddr=${BASH_REMATCH[1]}

[[ $(wwallet chain list | wc -l) == "0" ]]

# deploy a chain
wwallet chain deploy --chain=$alias --committee='0,1,2,3' --quorum=3
chainid=$(cat wwallet.json | jq .chains.$alias -r)

[[ $(wwallet chain list | wc -l) == "1" ]]
[[ $(wwallet chain list) =~ "$chainid" ]]

# unnecessary, since it is the latest deployed chain
wwallet set chain $alias

# test chain info command
r=$(wwallet chain info)
echo "$r"
# test that the chainid is shown
[[ "$r" =~ "$chainid" ]]

# test the list-contracts command
r=$(wwallet chain list-contracts)
echo "$r"
# check that root + accountsc + blob contracts are listed
[[ $(echo "$r" | wc -l) == "3" ]]

# test the list-accounts command
r=$(wwallet chain list-accounts)
echo "$r"
# check that the owner is listed
echo "$r" | grep -q "$owneraddr"

agentid=$(echo "$r" | grep "$owneraddr" | sed 's/[:[:space:]].*$//')

r=$(wwallet chain balance "$agentid")
echo "$r"
# check that the chain balance of owner is 1 IOTA
[[ "$r" == "IOTA: 1" ]]

# same test, this time calling the view function manually
r=$(wwallet chain call-view accounts-0.1 balance string a agentid "$agentid" | wwallet decode color int)
[[ "$r" == "IOTA: 1" ]]

echo "PASS"
