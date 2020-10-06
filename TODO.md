# TODO

Done
- [x] `fairroulette dashboard`: Add install instructions
- [x] `fairroulette dashboard`: Auto-refresh
- [x] `fairroulette dashboard`: Display SC address balance
- [x] deploy `FairRoulette` PoC
- [x] Release binaries
- [X] implement `FairAuction` smart contract with tests
- [x] Integration tests: end test when a specific message is published (instead
      of waiting for an arbitrary amount of seconds).
- [x] adjust WaspConn etc APIs to real Goshimmer APIs.
- [x] Extend wwallet with `FairAuction` and `FairRoulette`
- [x] dwf dashboard
- [x] dashboard: display SC hash/description/address/owner-address
- [x] `wwallet wallet init` -> `wwallet init`
- [x] wwallet: deploy generic SC from proghash + committee
- [x] deploy Wasp in Pollen testnet
- [x] deactivate/activate smart contract in the node
- [x] wasp node dashboard: show structure of committee, which SCs are running, etc

Pending
- [ ] wwallet: separate binaries for admin/client operations
- [ ] dwf: allow withdrawing colored tokens
- [ ] BufferedKVStore: Cache DB reads (which should not change in the DB during
      the BufferedKVStore lifetime)
- [ ] serialize access to solid state (ie, guarantee that state loaded with LoadSolidState does not
      change until released).
- [ ] Add authentication to web api calls

To discuss/RFC
- [ ] optimize SC ledger database. Currently, key/value is stored twice: in the virtual state and in the batch which
last updated the value. For small virtual states it is OK. For big ones (data Oracle) it would be better
to for virtual state keep reference to the last updating mutatation in the batch/state update 
- [ ] identity system for nodes
- [ ] secure access to API, to SC admin functions 
- [ ] refactor 'request code' from uint16 value to string
- [ ] smart contract state access from outside. The current approach is to provide universal node API to query state. 
The alternatives would be to expose access functions (like view in Solidity) from the smart contract code itself. 
Another approach can be expose data schema + generic access   
- [ ] Merkle proofs of smart contract state elements The idea is to have relatively short (logoarithmically) proof
of some data element is in the virtual state. Currently proof is the whole batch chain, i.e. linear.  
- [ ] Standard subscription mechanisms for events: (a) VM events (NanoMsg, ZMQ, MQTT) 
and (b) smart contract events (signalled by request to subscriber smart contract)
- [ ] balance sheet metaphor in the smart contract state. Ownership concept of BS "liability+equity" items  
- [ ] implement framework with mocked Sandbox for smart contract unit testing 
- [ ] "stealth" mode for request data. Option 1: encryption of it to committee members with symetric key encrypted
for each committee member with its public key. Option 2: move request data off-tangle and keep only hash of it on-tangle 

Functional testing
- [X] test fault-tolerance
- [ ] test access node function
- [X] test several concurrent/interacting contracts
- [X] test random confirmation delays (probably not needed if running on Pollen)
- [ ] test big committees (~100 nodes)
- [X] test overlapping committees

Future
- [ ] rewrite DKG
- [ ] `Oracle Data Bulletin Board` specs. Postponed
- [ ] enable and test 1 node committees
- [ ] test (or implement) quorum == 1  
- [ ] optimize logging
- [ ] Prometheus metrics
- [ ] MQTT publisher

# Roadmap
- `DonateWithFeedback`, `TokenRegistry`, `FairAuction`, `FairRoulette`,  on Pollen testnet
- Wasm VM and Rust environment for Wasm smart contracts. Programming tools. 
- PoC smart contracts on Rust->Wasm
- Node dashboard, smart contract dashboard, visualisation
- Admin tools and APIs 
- CLI wallet (universal and template)
- architecture WP 
