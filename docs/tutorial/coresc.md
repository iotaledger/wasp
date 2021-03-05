# Core smart contracts

The _virtual machine_ or _VM_ of the chain is a component responsible for the deterministic calculation of the next
state of the chain from the current state and requests. The input of the _VM_ task is an ordered batch of requests plus 
UTXO state of the chain address. The output of the VM task is the mutation of the chain state (the _block_) and 
anchor transaction, yet unsigned.   

One run of the _VM_ is represented by the _VMContext_ object. The _VMContext_ provides mutable context for the 
run of the batch by the smart contracts on the chain. It also contain access to smart contracts, deployed on the chain.

The are 4 core smart contracts always deployed on each chain. They ensure core logic of the VM and provide platform 
for plugging of other smart contracts into the chain: 
- [root](root.md) contract responsible for initialization of the chain, deployment of new contracts and other administrative 
fyunctions
- [blob](blob.md) contract responsible for on-chain register of arbitrary data _blobs_
- [accounts](accounts.md) contract is responsible for the system of on-chain accounts of colored tokens
- [eventlog](eventlog.md) contract is responsible for the on-chain event log  
