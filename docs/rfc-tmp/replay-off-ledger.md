# Replay protection for off-ledger requests

## Motivation

The _off-ledger requests_ (aka _L2 requests_) are ISCP requests sent to the smart contract directly
through the Wasp node, a validator node or access node of the chain. In contrast, _on-ledger requests_ (aka _L1 requests_)
are sent to the smart contract by wrapping the request into the IOTA value transaction and confirming it on the Tangle.

The _on-ledger requests_ thus are protected from replay because UTXO ledger on L1 ledger does not allow
confirming same transaction twice.

It is not the case with _off-ledger_ requests which are just data packets broadcasted among validator
and access nodes. The chain must prevent attacks when the same _off-ledger request_ is re-broadcasted.

## Replay protection by request receipts in the state

The VM through [blocklog core contract](../../packages/vm/core/blocklog/interface.go) stores ID of each request processed
by the chain in the state, together with the result information (receipt). It can be retrieved by its _request id_.
It is also the way to check if the request was already  processed in the past.

The [mempool](../../packages/chain/mempool/mempool.go#L116) checks the state and removes any requests which are already processed.

The above is valid for both _off-ledger_ and _on-ledger_ requests. It means current implementation is fundamentally protected  
against request replay.

However, the problem is with potentially huge logs. We are aiming at high TPS to be processed by the ISCP chains,
mostly by using mechanism of _off-ledger_ requests.

Let's assume we have 1000 TPS (requests per second). Each request generates a record 50+ bytes long in the `blocklog`.
This gives us 50 kB/s = 180 MB/h = 4.3 GB/day. This isn't sustainable in the longer run.

It means ISCP chain state will necessary be periodically pruned, deleting non-critical and obsolete information from the state.

_(In short, the pruning of the state involves extracting witnesses/proofs of existence of the past state
and then deleting that part from the state. Witnesses may be stored independently of the chain, while the chain will contain
root of the witness. Witnesses may be implemented by Merkle trees or other techniques, such as vector commitments)_

In any case, after some time it won't be possible to check in the state if the requests was processed in the past.

It creates and attack vector: the same request can be replayed by anyone after some time. This is unacceptable.

## Naive replay protection with increasing `nonce`

We may require from each request an ever-increasing value, the `nonce`. The VM would store the maximum value of the `nonce`
for requests with the same sender address. Then VM would reject any _off-ledger request_ from this address with the
`nonce` less or equal with the stored maximum. The use will be forced to use incrementally updated value for the `nonce` or
e.g. timestamp.

The good thing with this approach is that the check can be performed early by calling a view and checking the `blocklog`.

The problem, however, is that with our consensus we cannot guarantee the requests will be included into the block in
the order they arrived. For example, if a client sends 5 requests at once, with nonces `1, 2, 3, 4, 5`, the
requests may be picked and processed in two batches `1,2,5` and `3,4`.
In this case requests `3,4` will fail for no reason for the user.

## Proposed solution

We propose to combine the two methods: by logging and checking processed requests in the state and requiring
incremental nonce for _off-ledger requests_:

* each _off-ledger request_ is first checked in the (pruned) state if it wasn't processed before
* each _off-ledger request_ is required a _nonce_, `uint64` value
* _on-ledger requests_ don't have `nonce`
* for each new off-ledger request VM will store value `MaxAssumed` in the state next to the sender's address the following way:
  * if the `nonce` of the request is greather than existing `MaxAssumed`, it stores `nonce` as new `MaxAssumed`
  * otherwise it increments `MaxAssumed` by 1
* VM will validate any request the following way:
  * if `MaxAssumed` < `NConst` the request is **valid**
  * otherwise, if `nonce` > `MaxAssumed` - `NConst`, the requests **valid**
  * otherwise, the request is deemed **invalid**
* `NConst` is a global static constant, say `10000`.
* chain guarantee at least last `NConst` requests not pruned and present in the `blocklog` state for each address

`NConst` also known as `EnforceOrderNumberOfNoncesBack`. It must be at least the number of theoretically maximum number of  
_offledger requests_ from one address in one batch, i.e. with `NConst` number of requests in one batch the order is ignored.

The approach will enforce ever growing nonces. The user will be forced to use incremental nonces or use for example timestamp.  
VM will not allow too old `nonces` (too far in the past), but otherwise local sequence is not enforced and is assumed random.
