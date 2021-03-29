
# Chains

Registry of the `chain` objects. 

Contains all active chains the node is participating in.

There are `chains` package and `chains` plugin. 
* The package implements `Chains` object 
* The plugin creates one global instances of `Chains` and provides access to it

## Responsibility of the global `Chains` object

* handle the in-memory registry of chains: load from Registry ( `ChainRecord` ), activate, deactivate, find by `ChainID`
* attach to exposed events and listen to incoming async messages which are targeted to the specific chain:
  * `txstream`: transactions and other updates from Goshimmer
  * off-ledger requests from _off-ledger request module_
* dispatch messages to specific chain

# Chain object
Represents a chain the node is participating in.

The node can participate as:
* `committee node`. In this case it has its index and `DKShare` in a current committee of the chain. 
* `access node`. In this case it only contains state of the chain and cannot participate in producing state updates

## Requests

it is implemented as `coretypes.Request` interface

There are 2 types of requests which implements the interface:
* on-ledger requests which are coming as UTXO outputs of confirmed transactions from L1.
* off-ledger requests which are coming directly to the node through API.

### On-ledger requests
* corresponds to confirmed UTXO of `ledgerstate.ExtendedLockedOutput` type
* contains:
  * tokens (>=1)
  * target chain (`Address` of the output)
  * Metadata:
    * target contract
    * target entry point
    * params (RequestArgs, dict.Dict),
  * `sender`
  * `minted` proofs.
* coming with transaction from goshimmer via `txstream`
* each transaction is parsed into requests by adding `sender` and `minted` to each output
* RequestID == ledgerstate.OutputID == TransactionID||index

### Off-ledger requests
* coming as API calls from the wild. Do not correspond to any UTXO
* contains:
  * target chain (`ChainID`)
  * Metadata:
    * target contract
    * target entry point
    * params (dict.Dict),
  * `ordnum`: an increasing counter which must be unique for each transaction `uint64`
  * ED25519 signature of the above. Sender address is an address which can be derived from the signature
  * RequestID == ledgerstate.OutputID(blake2b(`contains`) || 0)
* each off-ledger request is checked:
  * signature must be valid
  * it must contain existing account on the chain for the sender address
  * its `ordnum` must be larger than the last `ordnum` stored with the account under the address

The off-ledger requests are validated upon receiving from API by:
* checking the signature
* calling a view to check if the account is present.
* checking if `ordnum` is higher than the one in the chain (it is only a preliminary check, to reject obviously wrong requests)

## Responsibility of the `Chain` object

* consume incoming stream of messages dispatched to the `chain` by `chains`.
* Message types:
  * Off-ledger requests, already pre-validated
  * Transactions from `txstream`. Each transaction is parsed:
    * `state transition`, the `AliasOutput` to the chain if it contains one
    * `on-ledger` requests, the `ExtendedLockedOutputs` with target address to the chain
* Process requests by producing blocks/updates of the chain's state
* maintain consistent solid state of the chain
* provide access to the solid state of the chain

### Peers

Peers of the chain are nodes which are running the chain.

Some peers are `committee peers`. The `committee peers` form a committee:
* committee represents `state address`, backed by a distributed key
* each committee peer contains index in the committee and the secret partial key, generated during DKG (must be present in the registry)

`Committee` may change, then `state address` and `committee peers` change too.

Peers are exchanging information (within the chain).
* receiving peer sends validated off-ledger requests to committee peers.
* each committee peer send `request id` of each request it is ready to process to the committee peers
* all peers exchange syncing data (blocks) upon request
* all peers exchange blob data
* committee peers exchange partial signatures of transactions

### Mempool

1. Committee peer maintains a pool of pending requests: the `mempool`. It contains all pre-validated and unprocessed yet requests.
    * Requests are placed into the `mempool` when they are coming from internal sources: as off-ledger requests and as requests parsed from transactions.
    * Requests are removed from the `mempool` when block is committed to the DB. The block contains IDs of all processed requests.
2. `mempool` taking care about solidification of `blob references`
3. `mempool` is producing a list of requests which node is _ready to process_ (taking into account of solidification
of blob references, timelocks and so on.
4. `mempool` is informing other committee peers and its own consensus process with its _ready to process_
list by sending corresponding request IDs.

### Consensus
* Receives _ready to process_ list, current `state output` and produces a signed transaction
* sends pending block and pending transaction to the `state manager`
* `Consensus` communicates with other committee nodes
* also `consensus` sends partial signatures to peers designated as `sentinels`

### State manager
* listens to the incoming state transitions (`AliasOutputs`)
* upon approval, commits block to DB
* when detects it is behind, sends pulls request to other peers to get missing block
* when needed, sends pull request for confirmed `state outputs`
* informs consensus about state transition

### Committee manager (?)
Handles committee transitions. Upon arrival of state output which is marked as _governance update_  
checks address of the current committee:
* if it is an access nodes wrt the chain, it tries to load `CommitteeRecord` and corresponding `DKShare`.
If success, sets up the committee and continue processing as committee node
* if it is a committee node wrt the chain and state address change, it tries to load `CommitteeRecord` and corresponding `DKShare`.
If success, sets up the new committee and continue processing as committee node
* if it is a committee node wrt the chain and address does not change, it does nothing
* if address change and node can't load the key, it continues as an access node


