# TODO

- [x] `fairroulette dashboard`: Add install instructions
- [x] `fairroulette dashboard`: Auto-refresh
- [x] `fairroulette dashboard`: Display SC address balance
- [x] deploy `FairRoulette` PoC
- [x] Release binaries
- [ ] Integration tests: end test when a specific message is published (instead
      of waiting for an arbitrary amount of seconds).
- [ ] BufferedKVStore: Cache DB reads (which should not change in the DB during
      the BufferedKVStore lifetime)
- [ ] implement framework with mocked Sandbox for smart contract unit testing 
- [X] implement `FairAuction` smart contract with tests
- [ ] adjust WaspConn etc APIs to real Goshimmer APIs.
- [ ] extend goshimmer cli-wallet with `FairAuction` and `FairRoulette`
- [ ] deploy Wasp in Pollen testnet
- [ ] `Oracle Data Bulletin Board` description

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
- Wasm VM and Rust environment for Wasm smart contracts 
- `FairRoulette` on Rust->Wasm
- `Oracle Data Bulletin Board` PoC
- Node dashboard, smart contract dashboard, visualisation
- Admin tools and APIs 
- CLI wallet (universal and template)
- architecture WP 
