# The `blocklog` contract

The `blocklog` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The function of the `blocklog` contract is to keep track of the blocks of
requests that were processed by the chain.

It provides functions to deposit and withdraw funds, also gives the count of
total assets deposited on the chain. Note that the ledger of accounts is
consistently maintained by the VM behind scenes,the `accounts`
core smart contract provides frontend of authorized access to those account by
outside users.

### Entry Points

The `blocklog` core contract does not contain any entry points which modify its
state.

The only way to modify the `blocklog` state is by submitting requests for
processing to the chain.

### Views

* **viewGetBlockInfo** - Returns the data of the block in the chain with
  specified index.

* **viewGetLatestBlockInfo** - Returns the index and data of the latest block in
  the chain.

* **viewGetRequestLogRecord** - Returns the data, block index, and request index
  of the specified request.

* **viewGetRequestLogRecordsForBlock** - Returns the data, block index, and
  request index of all requests in the block with the specified block index.

* **viewGetRequestIDsForBlock** - Returns the IDs of all requests in the block
  with the specified block index.

* **viewIsRequestProcessed** - Returns whether a request with specified ID has
  been processed.

