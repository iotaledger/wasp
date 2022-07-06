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

## Entry Points

The `blocklog` core contract does not contain any entry points which modify its
state.

The only way to modify the `blocklog` state is by submitting requests for
processing to the chain.

---

## Views

### - `getBlockInfo(n BlockIndex)`

Returns information about the block with index `n`. If `n` is not provided, it defaults to the current (latest) block.

Block info has the following data:

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

### - `getRequestIDsForBlock(n BlockIndex)`

Returns a list with the IDs of all requests in the block with block index `n`.

### - `getRequestReceipt()`

### - `getRequestReceiptsForBlock()`

### - `isRequestProcessed()`

Returns whether a request with specified ID has been processed.

### - `getEventsForRequest()`

Returns a list of events for a given request.

### - `getEventsForBlock()`

Returns a list of events for a given block.

### - `getEventsForContract()`

Returns a list of events for a given smart contract.
<!-- 
### `viewGetRequestLogRecord()`

Returns the data, block index, and request index of the specified request.

### `viewGetRequestLogRecordsForBlock()`

Returns the data, block index, and request index of all requests in the block with the specified block index.

 -->

### `viewccontrolAddresses()`
