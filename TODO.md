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

Pending
- [ ] wasp node dashboard: show structure of committee, which SCs are running, etc
- [ ] wwallet: Allow more than one instance of same SC
- [ ] dwf: allow withdrawing colored tokens
- [ ] BufferedKVStore: Cache DB reads (which should not change in the DB during
      the BufferedKVStore lifetime)
- [ ] serialize access to solid state (ie, guarantee that state loaded with LoadSolidState does not
      change until released).
- [ ] Add authentication to web api calls

To discuss/RFC
- [ ] refactor 'request code' from uint16 value to string
- [ ] smart contract state access from outside. The current approach is to provide universal node API to query state. 
The alternatives would be to expose access functions (like view in Solidity) from the smart contract code itself. 
Another approach can be expose data schema + generic access   
- [ ] Merkle proofs of smart contract state elements  
- [ ] Standard subscription mechanisms for events: (a) VM events (NanoMsg, ZMQ, MQTT) 
and (b) smart contract events (signalled by request to subscriber smart contract)
- [ ] balance sheet metaphor in the smart contract state. Ownership concept of BS "liability+equity" items  
- [ ] implement framework with mocked Sandbox for smart contract unit testing 

Functional testing
- [ ] test fault-tolerance
- [ ] test access node function
- [ ] test several concurrent/interacting contracts
- [ ] test random confirmation delays (probably not needed if running on Pollen)
- [ ] test big committees (~100 nodes)
- [ ] test overlapping committees

Future
- [ ] rewrite DKG
- [ ] `Oracle Data Bulletin Board` description. Postponed
- [ ] enable and test 1 node committees
- [ ] test quorum == 1  
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
