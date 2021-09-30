# Sandbox Interface

The only way a smart contract can interact with the world (access the state, call other smart contracts or send transactions) is by using the Sandbox interface.

The Sandbox provides limited and deterministic access to the state through a key/value storage abstraction.

![Sandbox](/img/sandbox.png)

Besides read/write to the contract state, the Sandbox interface allows Smart contracts to access :

- the AgentID of the contract
- the details of the current function invokation (request or view call)
- the balances owned by the contract
- the AgentID of whoever deployed the contract
- the timestamp of the current block
- cryptographic utilities (hashing, verify signatures, obtain addresses from public keys, etc)
- events dispatch
- Entropy (deterministic randomness)
- logging (usually only used for debugging when testing)
