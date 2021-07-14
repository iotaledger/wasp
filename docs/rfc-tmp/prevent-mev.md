# Prevention of MEV in ISCP

_MEV_ stand for _Miner Extractable Value_. Miner can take advantage of its knowledge and include and order its own transactions
into the block to its advantage: frontrunning, sandwitching, abusing early knowledge of oracle submitted values and similar.

It is a common thing e.g. in Ethereum because in Nakamoto PoW consensus miner defines the block content.

Generally speaking, in the BFT consensus used by ISCP it is impossible, because all validator node first agree on  
inputs to computations (the batch) and the rest is deterministic, i.e. cannot be influenced by less than 2/3+1 majority.

From the other side, ISCP consensus has to sort the batch deterministically before running the VM.  
This opens theoretical opportunity to take advantage if the order is known in advance.

To prevent it, we propose to use native unpredictable yet deterministic randomness generated durin the consensus  
run to sort the batch according to it.

Lets say R is a random number produced by the ACS component of the consensus. The we sort the batch for example by  
_(requestID+R) mod 2^16_. That will give deterministic yet unpredictable order in the batch of requests.  
MEV will be impossible because parties can't influence the order.

Note: this also means order in which requests are processed is random. This is fine because it is undetermined anyway.  
If user need strict order, it waits completion of the previous request before posting the next one.

