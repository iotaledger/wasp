# Replay protection for off-ledger requests

## Motivation

The _off-ledger requests_ (aka _L2 requests_) are ISCP requests sent to the smart contract directly
through the Wasp node, a validator node or access node of the chain. In contrast, _on-ledger requests_ (aka _L1 requests_)
are sent to the smart contract by wrapping the request into the IOTA value transaction and confirming it on the Tangle.

The _on-ledger requests_ thus are protected from replay because UTXO ledger does not allow confirming the same transaction twice.

This is not the case with _off-ledger_ requests which are just data packets broadcast among validator
and access nodes of the chain. Therefore, the chain must be equipped with a mechanism to prevent attacks when the
same _off-ledger request_ is re-posted again.

## Replay protection by request receipts in the state

The VM through [blocklog core contract](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/blocklog/interface.go) stores the ID of each request processed
by the chain in the state, together with the result information (receipt). It can be retrieved by its _request id_.
The _blocklog_ core contract also provides a way for fast check if the request was already processed in the past.

The [mempool](https://github.com/iotaledger/wasp/blob/develop/packages/chain/mempool/mempool.go) checks the state and removes any requests which are already processed.  
The processed request will never appear in the batch proposal during the consensus process.

The above is valid for both _off-ledger_ and _on-ledger_ requests. It means current implementation is fundamentally protected  
against request replay.

However, the problem is with potentially huge logs in the state (in the _blocklog_).  
The ISCP design is aiming at high TPS (settled requests per second) to be processed by the ISCP chains, mostly by using mechanism of _off-ledger_ requests.

Let's assume we have 1000 TPS (requests per second). Each request generates a record 50+ bytes long in the `blocklog`.
This gives us 50+ kB/s = 180+ MB/hour = 4.3+ GB/day. This isn't sustainable in the long run.

It means ISCP chain state will have to be equipped with the mechanism of periodical pruning, i.e. deleting non-critical and obsolete information from the state.

_(In short, the pruning of the state involves extracting witnesses/proofs of existence of the past state
and then deleting that part from the state. Witnesses may be stored independently of the chain, while the chain will contain
root of the witness, i.e. proof of existence of the witness.
Witnesses may be implemented by Merkle trees or by other techniques, such as polynomial KZG10 (vector) commitments)_

In any case, after some time it won't be possible to check in the state if the requests was processed in the past.

It creates an attack vector: the same _off-ledger request_ can be replayed by anyone after some time. This is unacceptable.

## Naive replay protection with increasing `nonce`

We may require from each request an ever-increasing value, the `nonce`. The VM would store the maximum value of the `nonce`
for requests with the same sender address. Then the VM would reject any _off-ledger request_ from this address with the
`nonce` less or equal with the stored maximum. The user will be forced to use incrementally updated value for the `nonce` or
e.g. timestamp. Once `nonce` is part of the essence of the request and is hashed into the _request id_, the replay becomes impossible.

The approach would allow an early check by calling a view and checking the `blocklog`. It could be used to prevent spamming/DDoS attacks.

The problem, however, is that with the ISCP consensus, we cannot guarantee the requests will be included in the block in
the order in which they arrived. Actually, the order is intentionally random.  
For example, if a client sends 5 requests at once, with correct sequence of nonces `1, 2, 3, 4, 5`, the requests may be
picked and processed in two separate batches `1,2,5` and `3,4`.
In this case requests `3,4` will fail for no reason for the user.

## Proposed solution

We propose to combine the two methods: by checking the processed requests in the state and at the same time requiring and enforcing an
incremental nonce for _off-ledger requests_:

* each _off-ledger request_ is first checked in the (pruned) state if it wasn't processed before
* each _off-ledger request_ is required a _nonce_, an `uint64` value. _On-ledger requests_ don't have a `nonce`
* for each new _off-ledger request_ the VM will keep a value `MaxAssumed` in the state next to the sender's address the following way:
  * if the `nonce` of the request is greater than the existing `MaxAssumed`, it stores `nonce` as new `MaxAssumed`
  * otherwise, it increments `MaxAssumed` by 1
* the VM will validate any request the following way:
  * if `MaxAssumed` < `NConst` the request is **valid**
  * otherwise, if `nonce` > `MaxAssumed` - `NConst`, the request is **valid**
  * otherwise, the request is deemed **invalid**
* `NConst` is a global static constant
* chain guarantee at least last `NConst` requests not pruned and therefore present in the `blocklog` state for each address

`NConst` also known as `OffLedgerNonceStrictOrderTolerance`. It must be at least be a theoretic maximum number of
_offledger requests_ from one address in one batch, i.e. with `NConst` number of requests in one batch the order is ignored.
The `10000` is a reasonable value.

The approach will enforce ever-growing `nonces`. The user will be forced to use incremental nonces or use for example a timestamp.
The VM will not allow too old `nonces` (too far in the past), but otherwise a local sequence is not enforced and is assumed to be random.
