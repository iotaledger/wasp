# Prevention of MEV in ISCP

_MEV_ stands for _Miner Extractable Value_. Miner can take advantage of its knowledge, include and order its own transactions
into the block to its advantage: front-running, sandwiching, abusing early knowledge of oracle submitted values and similar.

It is a common thing e.g. in Ethereum because in Nakamoto PoW consensus miner fully defines the block content.

Generally speaking, in the BFT consensus used by ISCP it is impossible. This is because the VM is run post-consensus.
All validator nodes first agree on inputs to computations (the _batch_), and the rest is deterministic,
i.e. cannot be influenced by less than 2/3+1 majority.

From the other side, in ISCP consensus the batch has to be sorted deterministically before running the VM. It opens a theoretical  
opportunity to take advantage in case the order or requests can be computed in advance.

To prevent it, we propose to use the native unpredictable yet deterministic randomness generated during the consensus.  
The batch will be sorted using that value to produce unpredictable order of requests.

Let's say R is a random number produced by the ACS component of the consensus and applying the BLS threshold cryptography.  
Then we can sort the batch for example by _(requestID+R) mod 2^16_.  
That will give a deterministic yet unpredictable order in the batch of requests.  
MEV will be impossible because parties won't be able to influence the order of requests in the batch.

Note: this also means that the order in which requests are processed is random for the user. This is fine because it is undefined anyway.  
If the user needs a strict order, he/she can wait for the completion of the previous request before posting the next one.

