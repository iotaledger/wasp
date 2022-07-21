### Gas usage tests

This folder holds gas usage tests for wasm. The solidity contracts can be found in
`wasp/packages/evm/evmtest` and corresponding tests in `packages/vm/core/evm/evmtest`. 

You can see current report at `./cmd`. You should see images for the different smart contracts `executiontime.png`, `storage.png`, and `memory.png`

### Generate report

A generated report is already attached. See images in `./cmd`. But if you make changes to the tests/contracts and want to generate a new report, follow these steps

1) Compile all modified contracts. See [wiki](https://wiki.iota.org/smart-contracts/guide/schema/usage)
2) Run tests. Running these tests will generate `.json` files. You need to run the wasm tests 3 times with `-gowasm`, `-tswasm` and `-rswasm` flags to genrate json output files
for Golang, TypeScript and Rust.
3) These json files should be created inside each contracts `pkg` folder. This folder is created when you compile the rust contract as instructed on the wiki 
4) Run `./cmd/main.go` to generate new images
