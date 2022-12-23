#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PARENT_DIR="$( builtin cd ${CURRENT_DIR}/.. >/dev/null 2>&1 ; pwd -P )"
cd ${PARENT_DIR}

LOOPS=10000
START=$(($(date +%s%N)/1000000))

COUNTER=0
while [  $COUNTER -lt $LOOPS ];do
    ./wasp-cli chain deposit 0x8B65DD08C7784017fe6B8Af20904e61916506fD4 base:100000 -w=false
    let COUNTER=COUNTER+1
done

END=$(($(date +%s%N)/1000000))
echo "DURATION: $(($END-$START))"

cd ${CURRENT_DIR}
