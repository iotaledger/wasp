# Core smart contracts

The _virtual machine_ or _VM_ of the chain is the component responsible for the
deterministic calculation of the next state of the chain from the current state
and requests. The input of the _VM_ task is an ordered batch of requests plus
UTXO state of the chain address. The output of the VM task is the mutation of
the chain state (the _block_) and an anchor transaction, yet unsigned.

One run of the _VM_ is represented by the _VMContext_ object. The _VMContext_
provides a mutable context for running the batch by the smart contracts on the
chain. It also provides access to the other smart contracts that are deployed on
the chain.

There are currently 6 core smart contracts that are always deployed on each
chain. They ensure core logic of the VM and provide a platform for plugging
other smart contracts into the chain:

- [root](root.md) contract responsible for initialization of the chain,
  deployment of new contracts, and other administrative functions.
- [_default](_default.md) catch-all contract for unhandled requests.
- [blob](blob.md) contract responsible for on-chain registration of arbitrary
  _data blobs_.
- [accounts](accounts.md) contract responsible for the ledger of on-chain
  colored token accounts.
- [blocklog](blocklog.md) contract keeps track of the blocks of requests that
  were processed by the chain.
- [eventlog](eventlog.md) contract responsible for the on-chain event log.  
