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
--- 
# The `blocklog` Contract

The `blocklog` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts chain.

The function of the `blocklog` contract is to keep track of the blocks of
requests that were processed by the chain.

It provides views to get request status or receipts, block information, or events (per request / block / smart contract).

## Entry Points

The `blocklog` core contract does not contain any entry points which modify its
state.

The only way to modify the `blocklog` state is by submitting requests for
processing to the chain.

## Views

### viewGetBlockInfo

Returns the data of the block in the chain with specified index.

### viewGetLatestBlockInfo

Returns the index and data of the latest block in the chain.

### viewGetRequestLogRecord

Returns the data, block index, and request index of the specified request.

### viewGetRequestLogRecordsForBlock

Returns the data, block index, and request index of all requests in the block with the specified block index.

### viewGetRequestIDsForBlock

Returns the IDs of all requests in the block with the specified block index.

### viewIsRequestProcessed

Returns whether a request with specified ID has been processed.

### viewGetEventsForRequest

Returns a list of events for a given request.

### viewGetEventsForBlock

Returns a list of events for a given block.
  
### viewGetEventsForContract

Returns a list of events for a given smart contract.
  