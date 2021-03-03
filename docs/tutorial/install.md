## Installation (TBD)

_Solo_ tests are written in Go. The tutorial assumes you are familiar with and have installed Go 1.15 
environment. The _Solo_ package can be installed by cloning the Wasp repository:
```
git clone https://github.com/iotaledger/wasp.git
```

Note: the Solo package can be installed separately using this command:
```
go get github.com/iotaledger/wasp/packages/solo
```

In Windows, use: 
```
go get buildmode=exe github.com/iotaledger/wasp/packages/solo
```

To run Rust/Wasm smart contracts you will need Rust and wasm-pack installed. 
You can use any development environment for Rust and Go. 
The GoLand environment with the Rust plugin is a good combination.

You can find example implementations of smart contracts (including source code
and tests) in the Wasp repository,
[`contracts/rust`](https://github.com/iotaledger/wasp/tree/master/contracts/rust)
folder.
