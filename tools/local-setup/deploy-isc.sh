IOTA_ADDRESS=$(iota keytool list --json | jq -r '.[0].iotaAddress')
# Load the modules field from bytecode.json and format as JSON array
COMPILED_MODULES=$(jq -c '.modules' /isc/bytecode.json)

DEPENDENCIES=$(jq -c '.dependencies' /isc/bytecode.json)
GAS_BUDGET=100000000

# Make the JSON-RPC POST call
PUBLISH_RAW_UNSIGNED_TX_BYTES=$(curl -X POST http://host.docker.internal:9000 \
  -H "Content-Type: application/json" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 1,
    \"method\": \"unsafe_publish\",
    \"params\": [
      \"$IOTA_ADDRESS\",
      $COMPILED_MODULES,
      $DEPENDENCIES,
      "null",
      \"$GAS_BUDGET\"
    ]
  }")

UNSIGNED_TX_BYTES=$(echo "$PUBLISH_UNSIGNED_TX_BYTES" | jq -r '.result.txBytes')
SERIALIZED_SIGNATURE=$(iota keytool sign --address $IOTA_ADDRESS --data $UNSIGNED_TX_BYTES --json | jq -r '.iotaSignature')
RAW_PUBLISH_RESULT=$(iota client execute-signed-tx --tx-bytes $UNSIGNED_TX_BYTES --signatures $SERIALIZED_SIGNATURE  --json)
PACKAGE_ID=$(echo "$RAW_PUBLISH_RESULT" | jq -r '.objectChanges[] | select(.type == "published") | .packageId')

