#!/bin/bash

ENVS=$(sui client envs | grep '*')
VALIDATOR_PID=$(pidof -x sui-test-validator)
FAUCET_REQUESTS=25
FAUCET_WAIT_TIME="10s"

function exit_early() {
  echo "This script requires 'sui-test-validator' to be running"
  echo "and 'sui client' to be pointing to the local node (localhost:9000)"
  exit
}

if [[ -z $VALIDATOR_PID ]]; then
  echo "(test validator not running)"
  exit_early
fi

if [[ $ENVS != *"127.0.0.1"* && $ENVS != *"localhost"* ]]; then
  echo "(Environment wrong)"
  exit_early
fi

if [ $FAUCET_REQUESTS -gt 0 ]; then
  for i in $(seq $FAUCET_REQUESTS); do
    sui client faucet
  done


  echo "Waiting $FAUCET_WAIT_TIME for funds to arive."
  sleep $FAUCET_WAIT_TIME
fi

cd contract/isc

sui move build
sui client publish --gas-budget 5000000000  

cd ../../