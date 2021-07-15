# TODO

- [ ] gas and/or time budgets for VM entry point calls
- [ ] wasp-cli: separate binaries for admin/client operations
- [ ] dwf: allow withdrawing colored tokens
- [ ] BufferedKVStore: Cache DB reads (which should not change in the DB during
      the BufferedKVStore lifetime)
- [ ] serialize access to solid state (ie, guarantee that state loaded with LoadSolidState does not
      change until released).
- [ ] Add authentication to web api calls. Done ??
- [ ] discuss market for iota/colored coins + trustless oracle for every chain

### To discuss/RFC
-  [ ] accounts and other core contracts don't need tokens. 
    Possible policy: if caller is a core contract, accrue it all to the chain owner
- [ ] optimize SC ledger database. Currently, key/value is stored twice: in the virtual state and in the batch which
last updated the value. For small virtual states it is OK. For big ones (data Oracle) it would be better
to for virtual state keep reference to the last updating mutatation in the batch/state update 
- [ ] identity system for nodes
- [ ] (Merkle) proofs of smart contract state elements The idea is to have relatively short (logoarithmically) proof
of some data element is in the virtual state. Currently proof is the whole batch chain, i.e. linear.  
- [ ] Standard subscription mechanisms for events: (a) VM events (NanoMsg, ZMQ, MQTT) 
and (b) smart contract events (signalled by request to subscriber smart contract)
- [ ] "stealth" mode for request data. Option 1: encryption of it to committee members with symetric key encrypted
for each committee member with its public key. Option 2: move request data off-tangle and keep only hash of it on-tangle 

### Functional testing
- [ ] test access node function
- [ ] test big committees (~100 nodes)

### Nice to have
- [ ] Prometheus metrics
- [ ] MQTT publisher
- [ ] `Oracle Data Bulletin Board` specs. Postponed

## ISCP Core beta. 2Q 2021 (not finished)
- Wasp beta version on Nectar (optionally, on Chrysalis)
- release 2 ISCP Core Architecture specs  
- Core BFT consensus vetted and peer reviewed. Adjusted to Nectar version of the underlying ledger
- Merkle proofs of inclusion into the state
- identity system for nodes, node owners and SC owners
- complete committee change protocol based on ColorLockedOutputs 
- Ver 2 SC development tools, libraries and tutorials/docs for Rust 
- Ver 2 SC client libraries for Go, Rust and Javascript
