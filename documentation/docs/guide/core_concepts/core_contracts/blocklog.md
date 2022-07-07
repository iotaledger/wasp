---
description: The `blocklog` contract is to keep track of the blocks of requests that were processed by the chain. It also provides views to get request status, receipts, block information, or events.
image: /img/logo/WASP_logo_dark.png
keywords:
- core contracts
- blocklog
- views
- information
- request status
- receipts
- events
- reference
--- 
# The `blocklog` Contract

The `blocklog` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts chain.

The function of the `blocklog` contract is to keep track of the blocks of
requests that were processed by the chain.

It provides views to get request status or receipts, block information, or events (per request / block / smart contract).

---

## Block Information

```go
 BlockIndex                uint32
 Timestamp                 Time
 TotalRequests             uint16
 NumSuccessfulRequests     uint16
 NumOffLedgerRequests      uint16
 PreviousL1Commitment      Hash
 L1Commitment              Hash     
 AnchorTransactionID       TransactionID  
 TransactionSubEssenceHash Hash
 TotalIotasInL2Accounts    uint64
 TotalDustDeposit          uint64
 GasBurned                 uint64
 GasFeeCharged             uint64
```

---

## Request Receipt

```go
 Request       ISC Request
 Error         Unresolved VM Error 
 GasBudget     uint64                  
 GasBurned     uint64                  
 GasFeeCharged uint64                  
 BlockIndex   uint32       
 RequestIndex uint16       
```

:::note
Errors on receipts queried directly from blocklog are not humanly readable.

Those errors need to be translated using // TODO add link
:::

---

## Entry Points

The `blocklog` core contract does not contain any entry points which modify its
state.

The only way to modify the `blocklog` state is by submitting requests for
processing to the chain.

---

## Views

### - `getBlockInfo(n BlockIndex)`

Returns information about the block with index `n`. If `n` is not provided, it defaults to the current (latest) block.

### - `getRequestIDsForBlock(n BlockIndex)`

Returns a list with the IDs of all requests in the block with block index `n`.

### - `getRequestReceipt(u RequestID)`

Returns the receipt for a request with ID `u`.

### - `getRequestReceiptsForBlock(n BlockIndex)`

Returns all the receipt for the block with index `n`.

### - `isRequestProcessed(u RequestID)`

Returns whether a request with ID `u` has been processed.

### - `getEventsForRequest(u RequestID)`

Returns a list of events for a request with ID `u`.

### - `getEventsForBlock(n blockIndex)`

Returns a list of events for a block with index `n`.

### - `getEventsForContract(h Hname)`

Returns a list of events for a smart contract with hname `h`.

### `controlAddresses()`

Returns the current "State Controller", "Governing Address" and at what BlockIndex those were set.
