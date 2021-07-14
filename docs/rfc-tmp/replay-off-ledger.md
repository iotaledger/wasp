# Replay protection for off-ledger requests

## Motivation

The _off-ledger requests_ (aka _L2 requests_) are ISCP requests sent to the smart contract directly
through the Wasp node, a validator node or access node of the chain. In contrast, _on-ledger requests_ (aka _L1 requests_)  
are sent to the smart contract by wrapping the request into the IOTA value transaction and confirming it on the Tangle.

The _on-ledger requests_ thus are protected from replay because UTXO ledger on L1 ledger does not allow
confirming same transaction twice.

It is not the case with _off-ledger_ requests which are just data packets which are broadcasted among validators
and access nodes. The chain must prevent situations when the same _off-ledger request_ is rebroadcasted.

## Replay protection by request receipts

The VM through [blocklog core contract](../../packages/vm/core/blocklog/interface.go) stores each request processed
by the chain in the state, together with the result information (receipt). It can be retrieved by its _request id_.  
It is also the way of checking if a request wasalready  processed in the past.

The [mempool](../../packages/chain/mempool/mempool.go#L116) checks the state and removes any requests which are processed.

The above is valid for both _off-ledger_ and _on-ledger_ requests. It means current implementation is fundamentally protected  
against request replay.

The problem is we are aiming at high TPS to be processed by the ISCP chains, mostly by using mechanism of _off-ledger_ requests.  
Let's assume 1000 TPS (requests per second). Each request generates a record at least 50 bytes long in the `blocklog`.
This gives us 50 kB/s = 180 MB/h = 4.3 GB/day. This isn't sustainable.

It means ISCP chain state will necessary be periodically pruned, deleting non-critical information from the state.  
(In short, the pruning of the state involves extracting witnesses (proofs of existence)  
of the past state and the deleting that part of the state.