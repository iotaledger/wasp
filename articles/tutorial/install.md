## Installation (TBD)

_Solo_ tests are written in Go. The tutorial assumes you are familiar with and have installed Go 1.15 
environment. The _Solo_ package can be installed by cloning the Wasp repository, default `develop` branch: 
```
git clone https://github.com/iotaledger/wasp.git
```

The Solo package can be installed separately using this command:
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

The source code for the example smart contracts and tests 
can be found in `wasplib` [repository](https://github.com/iotaledger/wasplib/tree/develop/rust/contracts/examples/example1)
