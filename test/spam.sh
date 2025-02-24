#!/bin/bash

# Check if an argument is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <address-index>"
    exit 1
fi

# Store the address index argument
addr_index=$1

# Default value for n (how often to call request-funds)
n=200
counter=0

wasp-cli request-funds --address-index=$addr_index

while true; do
    if [ $((counter % n)) -eq 0 ] && [ $counter -ne 0 ]; then
        echo "Requesting funds... (Address Index: $addr_index)"
        wasp-cli request-funds --address-index=$addr_index
    else
        echo "Depositing... (Address Index: $addr_index)"
        wasp-cli chain deposit 0xD2598952a2579818983807dC188fDF7384d6F6CA "base|111" --node=wasp1 --address-index=$addr_index
    fi

    counter=$((counter + 1))

    # Optional: Add a small delay to prevent overwhelming the system
    sleep 0.15
done