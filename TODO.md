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

Pending
- [ ] BufferedKVStore: Cache DB reads (which should not change in the DB during
      the BufferedKVStore lifetime)
- [ ] serialize access to solid state (ie, guarantee that state loaded with LoadSolidState does not
      change until released).
- [ ] implement framework with mocked Sandbox for smart contract unit testing 
- [ ] deploy Wasp in Pollen testnet
- [ ] `Oracle Data Bulletin Board` description

To discuss
- [ ] smart contract state access from outside. The current approach is to provide universal node API to query state. 
The alternative would be to expose access functions (view in Solidity) from the smart contract code itself.
- [ ] Merkle proofs of smart contract state elements  
- [ ] Standard subscription mechanisms for events: (a) VM events (NanoMsg, ZMQ, MQTT) 
and (b) smart contract events (signalled by request to subscriber smart contract) 

Functional testing
- [ ] test access node function
- [ ] test several concurrent/interacting contracts
- [ ] test random confirmation delays (probably not needed if running on Pollen)
- [ ] test big committees (~100 nodes)
- [ ] test overlapping committees

Future
- [ ] enable and test 1 node committees
- [ ] test quorum == 1  
- [ ] optimize logging
- [ ] Prometheus metrics
- [ ] MQTT publisher

# Roadmap
- `FairRoulette` and `FairAuction` on Goshimmer testnet
- Wasm VM and Rust environment for Wasm smart contracts. Programming tools. 
- `FairRoulette`, `FairAuction` and `Oracle Data Bulletin Board` PoC on Rust->Wasm
- Node dashboard, smart contract dashboard, visualisation
- Admin tools and APIs 
- CLI wallet (universal and template)
- architecture WP 
