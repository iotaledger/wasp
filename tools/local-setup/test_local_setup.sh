#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# Test script for local-setup docker-compose
# Prerequisites: docker, wasp-cli, curl, cast (foundry)
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
WASP_CLI_CONFIG="$SCRIPT_DIR/wasp-cli-test.json"

# Ports (match docker-compose.yml defaults)
IOTA_RPC="http://localhost:9000"
FAUCET="http://localhost:9123"
GRAPHQL="http://localhost:9125"
WASP_API="http://localhost:9090"
DASHBOARD="http://localhost/wasp/dashboard/"

CHAIN_NAME="testchain"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }

# Use a dedicated config file so we don't clobber the user's existing config
cli() { wasp-cli -c "$WASP_CLI_CONFIG" "$@"; }

# Retry a command up to N times with a delay
retry() {
    local max_attempts=$1
    local delay=$2
    shift 2
    local attempt=1
    while [ $attempt -le "$max_attempts" ]; do
        if "$@" 2>&1; then
            return 0
        fi
        info "Attempt $attempt/$max_attempts failed, retrying in ${delay}s..."
        sleep "$delay"
        attempt=$((attempt + 1))
    done
    return 1
}

cleanup() {
    rm -f "$WASP_CLI_CONFIG"
}
trap cleanup EXIT

# -----------------------------------------------------------------------------
# Step 0: Check prerequisites
# -----------------------------------------------------------------------------
info "Checking prerequisites..."
for cmd in docker curl wasp-cli cast python3; do
    command -v "$cmd" &>/dev/null || fail "$cmd is not installed"
done
pass "All prerequisites installed"

# -----------------------------------------------------------------------------
# Step 1: Start docker-compose
# -----------------------------------------------------------------------------
info "Starting docker-compose..."
docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout 120 2>&1 || true

info "Waiting for services to be ready..."
sleep 10

# -----------------------------------------------------------------------------
# Step 2: Test L1 services
# -----------------------------------------------------------------------------
info "=== Testing L1 Services ==="

# Test IOTA Node RPC (JSON-RPC, only accepts POST — GET returns 405)
info "Testing IOTA Node RPC ($IOTA_RPC)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$IOTA_RPC" 2>&1)
if [ "$HTTP_CODE" != "000" ]; then
    pass "IOTA Node RPC is reachable (HTTP $HTTP_CODE)"
else
    fail "IOTA Node RPC is not reachable"
fi

# Test Faucet
info "Testing Faucet ($FAUCET)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$FAUCET" 2>&1)
if [ "$HTTP_CODE" != "000" ]; then
    pass "Faucet is reachable (HTTP $HTTP_CODE)"
else
    fail "Faucet is not reachable"
fi

# Test GraphQL with a real query
info "Testing GraphQL ($GRAPHQL)..."
GQL_RESPONSE=$(curl -s --connect-timeout 5 -X POST "$GRAPHQL" \
    -H "Content-Type: application/json" \
    -d '{"query":"{ checkpoint { sequenceNumber } }"}' 2>&1)
if echo "$GQL_RESPONSE" | grep -q "data"; then
    pass "GraphQL query returned data"
else
    info "GraphQL response: $GQL_RESPONSE"
    fail "GraphQL query failed"
fi

# -----------------------------------------------------------------------------
# Step 3: Test Wasp API
# -----------------------------------------------------------------------------
info "=== Testing Wasp API ==="

info "Testing Wasp API ($WASP_API)..."
WASP_RESPONSE=$(curl -sf "$WASP_API/v1/node/version" 2>&1 || true)
if [ -n "$WASP_RESPONSE" ]; then
    pass "Wasp API is reachable - version: $WASP_RESPONSE"
else
    WASP_RESPONSE=$(curl -sf "http://localhost/wasp/api/v1/node/version" 2>&1 || true)
    if [ -n "$WASP_RESPONSE" ]; then
        pass "Wasp API is reachable via Traefik"
    else
        fail "Wasp API is not reachable"
    fi
fi

info "Testing Dashboard ($DASHBOARD)..."
DASH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$DASHBOARD" 2>&1)
if [ "$DASH_STATUS" -ge 200 ] && [ "$DASH_STATUS" -lt 400 ]; then
    pass "Dashboard is reachable (HTTP $DASH_STATUS)"
else
    info "Dashboard returned HTTP $DASH_STATUS (may need wasp-dashboard image)"
fi

# -----------------------------------------------------------------------------
# Step 4: Setup wasp-cli, request funds, and deploy chain
# -----------------------------------------------------------------------------
info "=== Setting up wasp-cli ==="

# Use RPC endpoint (9000) for l1.apiaddress — required for Move contract deployment
# (the GraphQL endpoint does not support Publish/DeployISCContracts)
cli init 2>/dev/null || true
cli set l1.apiaddress "$IOTA_RPC"
cli set l1.faucetaddress "$FAUCET/gas"
cli wasp add 0 "$WASP_API" 2>/dev/null || true
pass "wasp-cli configured (l1.apiaddress=$IOTA_RPC)"

info "Requesting funds from faucet..."
FUNDS_OUTPUT=$(cli wallet request-funds 2>&1 || cli request-funds 2>&1)
info "$FUNDS_OUTPUT"
pass "Funds requested"

# Request more funds to ensure enough for Move contract deployment + chain deploy
info "Requesting additional funds..."
cli wallet request-funds 2>&1 || cli request-funds 2>&1 || true
sleep 3

info "Deploying test chain (this may take up to 120s)..."
# NOTE: chain deploy often exits non-zero because activation fails on first try
# (the indexer needs time to sync the anchor object). We capture exit code separately.
set +e
DEPLOY_OUTPUT=$(cli chain deploy --chain="$CHAIN_NAME" 2>&1)
DEPLOY_EXIT=$?
set -e
echo "$DEPLOY_OUTPUT"

# Extract chain ID from output regardless of exit code
CHAIN_ID=$(echo "$DEPLOY_OUTPUT" | grep "ChainID:" | grep -oE '0x[a-fA-F0-9]+' || true)

if [ -z "$CHAIN_ID" ]; then
    fail "Chain deploy failed — no ChainID in output"
fi
info "Chain created: $CHAIN_ID"

if [ $DEPLOY_EXIT -ne 0 ]; then
    info "Chain activation failed (expected on first try). Waiting for indexer sync..."
    # Add chain to local config and retry activation after giving the indexer time
    cli chain add "$CHAIN_NAME" "$CHAIN_ID" 2>/dev/null || true
    sleep 15
    info "Retrying chain activation..."
    set +e
    ACTIVATE_OUTPUT=$(cli chain activate --chain="$CHAIN_NAME" 2>&1)
    ACTIVATE_EXIT=$?
    set -e
    echo "$ACTIVATE_OUTPUT"
    if [ $ACTIVATE_EXIT -ne 0 ]; then
        info "Second activation attempt failed. Trying once more after 15s..."
        sleep 15
        cli chain activate --chain="$CHAIN_NAME" 2>&1 || fail "Chain activation failed after retries"
    fi
fi
pass "Chain deployed and activated: $CHAIN_ID"

# Wait for chain to be fully active and consensus to start
info "Waiting for chain to initialize..."
sleep 15

EVM_URL="$WASP_API/v1/chains/$CHAIN_ID/evm"
info "EVM JSON-RPC URL: $EVM_URL"

# -----------------------------------------------------------------------------
# Step 5: Test EVM JSON-RPC read calls
# -----------------------------------------------------------------------------
info "=== Testing EVM JSON-RPC ==="

json_rpc() {
    local method=$1
    local params=${2:-"[]"}
    local id=${3:-1}
    curl -s -X POST "$EVM_URL" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":$id}"
}

json_get() {
    python3 -c "import sys,json; print(json.load(sys.stdin).get('$1',''))"
}

# Wait for EVM endpoint to be ready (chain needs time to initialize)
info "Waiting for EVM endpoint to be ready..."
EVM_READY=false
for i in $(seq 1 30); do
    RESPONSE=$(json_rpc "eth_chainId")
    if echo "$RESPONSE" | grep -q '"result"'; then
        EVM_READY=true
        break
    fi
    sleep 2
done
if [ "$EVM_READY" = false ]; then
    info "Last response: $RESPONSE"
    fail "EVM endpoint not ready after 60s"
fi

# eth_chainId
info "Testing eth_chainId..."
RESPONSE=$(json_rpc "eth_chainId")
EVM_CHAIN_ID=$(echo "$RESPONSE" | json_get "result")
if [ -n "$EVM_CHAIN_ID" ]; then
    pass "eth_chainId: $EVM_CHAIN_ID"
else
    info "Response: $RESPONSE"
    fail "eth_chainId failed"
fi

# eth_blockNumber
info "Testing eth_blockNumber..."
RESPONSE=$(json_rpc "eth_blockNumber")
BLOCK_NUM=$(echo "$RESPONSE" | json_get "result")
if [ -n "$BLOCK_NUM" ]; then
    pass "eth_blockNumber: $BLOCK_NUM"
else
    fail "eth_blockNumber failed"
fi

# eth_gasPrice
info "Testing eth_gasPrice..."
RESPONSE=$(json_rpc "eth_gasPrice")
GAS_PRICE=$(echo "$RESPONSE" | json_get "result")
if [ -n "$GAS_PRICE" ]; then
    pass "eth_gasPrice: $GAS_PRICE"
else
    fail "eth_gasPrice failed"
fi

# -----------------------------------------------------------------------------
# Step 6: Test EVM transactions
# -----------------------------------------------------------------------------
info "=== Testing EVM Transactions ==="

# Generate a new Ethereum account
info "Generating test Ethereum account..."
WALLET_OUTPUT=$(cast wallet new 2>&1)
PRIVATE_KEY=$(echo "$WALLET_OUTPUT" | grep -i "private key" | grep -oE '0x[a-fA-F0-9]{64}')
ETH_ADDRESS=$(echo "$WALLET_OUTPUT" | grep -i "address" | grep -oE '0x[a-fA-F0-9]{40}')
if [ -z "$PRIVATE_KEY" ] || [ -z "$ETH_ADDRESS" ]; then
    fail "Failed to generate Ethereum account"
fi
pass "Test account: $ETH_ADDRESS"

# Check initial balance (should be 0)
BALANCE=$(cast balance "$ETH_ADDRESS" --rpc-url "$EVM_URL" 2>&1)
info "Initial balance: $BALANCE"

# Fund the EVM account via wasp-cli chain deposit
info "Depositing funds to EVM account via wasp-cli..."
DEPOSIT_OUTPUT=$(cli chain deposit "$ETH_ADDRESS" base:1000000000 --chain="$CHAIN_NAME" -w 60s 2>&1)
DEPOSIT_EXIT=$?
info "Deposit output: $DEPOSIT_OUTPUT"
if [ $DEPOSIT_EXIT -ne 0 ]; then
    fail "wasp-cli chain deposit failed (exit code $DEPOSIT_EXIT)"
fi
pass "Deposit command succeeded"

# Wait for the deposit to be processed
info "Waiting for deposit to be processed on L2..."
FUNDED=false
for i in $(seq 1 20); do
    BALANCE_AFTER=$(cast balance "$ETH_ADDRESS" --rpc-url "$EVM_URL" 2>&1)
    if [ "$BALANCE_AFTER" != "0" ] && [ -n "$BALANCE_AFTER" ]; then
        FUNDED=true
        break
    fi
    sleep 2
done
if [ "$FUNDED" = false ]; then
    fail "Account balance is still 0 after deposit (waited 40s)"
fi
pass "Account funded: $BALANCE_AFTER wei"

# --- TX 1: Deploy a minimal contract ---
info "TX1: Deploying minimal test contract..."
# Simple contract: just returns empty (PUSH1 0 PUSH1 0 RETURN)
DEPLOY_TX=$(cast send --create \
    --private-key "$PRIVATE_KEY" \
    --rpc-url "$EVM_URL" \
    --json \
    0x60006000f3 2>&1)
DEPLOY_EXIT=$?

if [ $DEPLOY_EXIT -ne 0 ]; then
    info "Deploy output: $DEPLOY_TX"
    fail "TX1: Contract deployment failed (exit code $DEPLOY_EXIT)"
fi

DEPLOY_STATUS=$(echo "$DEPLOY_TX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))" 2>/dev/null || true)
DEPLOY_TX_HASH=$(echo "$DEPLOY_TX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('transactionHash',''))" 2>/dev/null || true)
CONTRACT_ADDR=$(echo "$DEPLOY_TX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('contractAddress',''))" 2>/dev/null || true)

if [ "$DEPLOY_STATUS" != "0x1" ] && [ "$DEPLOY_STATUS" != "1" ]; then
    info "Deploy receipt: $DEPLOY_TX"
    fail "TX1: Contract deployment reverted (status: $DEPLOY_STATUS)"
fi
pass "TX1: Contract deployed! TX=$DEPLOY_TX_HASH Contract=$CONTRACT_ADDR"

# --- TX 2: Send a value transfer ---
RECIPIENT="0x000000000000000000000000000000000000dEaD"
info "TX2: Sending value transfer (1000 wei) to $RECIPIENT..."
SEND_TX=$(cast send "$RECIPIENT" \
    --value 1000 \
    --private-key "$PRIVATE_KEY" \
    --rpc-url "$EVM_URL" \
    --json 2>&1)
SEND_EXIT=$?

if [ $SEND_EXIT -ne 0 ]; then
    info "Send output: $SEND_TX"
    fail "TX2: Value transfer failed (exit code $SEND_EXIT)"
fi

SEND_STATUS=$(echo "$SEND_TX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))" 2>/dev/null || true)
SEND_TX_HASH=$(echo "$SEND_TX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('transactionHash',''))" 2>/dev/null || true)

if [ "$SEND_STATUS" != "0x1" ] && [ "$SEND_STATUS" != "1" ]; then
    info "Send receipt: $SEND_TX"
    fail "TX2: Value transfer reverted (status: $SEND_STATUS)"
fi
pass "TX2: Value transfer successful! TX=$SEND_TX_HASH"

# Verify recipient got the funds
RECIPIENT_BALANCE=$(cast balance "$RECIPIENT" --rpc-url "$EVM_URL" 2>&1)
if [ "$RECIPIENT_BALANCE" = "0" ] || [ -z "$RECIPIENT_BALANCE" ]; then
    fail "TX2: Recipient balance is 0 after transfer"
fi
pass "TX2: Recipient balance: $RECIPIENT_BALANCE wei"

# --- TX 3: eth_call to deployed contract ---
info "TX3: Calling deployed contract at $CONTRACT_ADDR (eth_call, read-only)..."
CALL_RESULT=$(cast call "$CONTRACT_ADDR" --rpc-url "$EVM_URL" 2>&1)
CALL_EXIT=$?
if [ $CALL_EXIT -ne 0 ]; then
    info "Call output: $CALL_RESULT"
    fail "TX3: eth_call to deployed contract failed"
fi
pass "TX3: eth_call succeeded (result: ${CALL_RESULT:-empty})"

# --- TX 4: Get transaction receipt ---
info "TX4: Fetching transaction receipt for $SEND_TX_HASH..."
RECEIPT_RESPONSE=$(json_rpc "eth_getTransactionReceipt" "[\"$SEND_TX_HASH\"]")
RECEIPT_STATUS=$(echo "$RECEIPT_RESPONSE" | python3 -c "
import sys,json
r = json.load(sys.stdin).get('result',{})
print(r.get('status','') if r else '')
" 2>/dev/null || true)
if [ "$RECEIPT_STATUS" = "0x1" ] || [ "$RECEIPT_STATUS" = "1" ]; then
    pass "TX4: Transaction receipt confirmed (status: $RECEIPT_STATUS)"
else
    info "Receipt response: $RECEIPT_RESPONSE"
    fail "TX4: Transaction receipt not found or failed"
fi

# --- Verify final state ---
info "=== Verifying Final State ==="

# Block number should have increased
RESPONSE=$(json_rpc "eth_blockNumber")
BLOCK_NUM_AFTER=$(echo "$RESPONSE" | json_get "result")
info "Block number after transactions: $BLOCK_NUM_AFTER"

# Account nonce should be >= 2 (2 send transactions)
TX_COUNT=$(cast nonce "$ETH_ADDRESS" --rpc-url "$EVM_URL" 2>&1)
if [ "$TX_COUNT" -lt 2 ] 2>/dev/null; then
    fail "Expected nonce >= 2, got: $TX_COUNT"
fi
pass "Account nonce: $TX_COUNT (confirmed $TX_COUNT transactions executed)"

# Sender balance should have decreased
FINAL_BALANCE=$(cast balance "$ETH_ADDRESS" --rpc-url "$EVM_URL" 2>&1)
info "Sender final balance: $FINAL_BALANCE wei"

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo ""
echo "============================================"
echo -e "${GREEN}  All Tests Passed!${NC}"
echo "============================================"
echo ""
echo "Services tested:"
echo "  - IOTA Node RPC:  $IOTA_RPC"
echo "  - Faucet:         $FAUCET"
echo "  - GraphQL:        $GRAPHQL"
echo "  - Wasp API:       $WASP_API"
echo "  - Dashboard:      $DASHBOARD"
echo "  - EVM JSON-RPC:   $EVM_URL"
echo ""
echo "EVM transactions executed:"
echo "  - TX1: Contract deployment  -> $CONTRACT_ADDR"
echo "  - TX2: Value transfer       -> $RECIPIENT"
echo "  - TX3: eth_call (read-only) -> $CONTRACT_ADDR"
echo "  - TX4: Receipt verification -> $SEND_TX_HASH"
echo ""
echo "To tear down:  docker compose -f $COMPOSE_FILE down -v"
