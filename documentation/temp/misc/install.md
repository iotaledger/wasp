# Installation (TBD)

_Solo_ tests are written in Go. The tutorial assumes you are familiar with and
have installed the Go 1.16 environment. The _Solo_ package can be installed by
cloning the Wasp repository:

```
git clone https://github.com/iotaledger/wasp.git
```

Note: the Solo package can be installed separately using this command:

```
go get github.com/iotaledger/wasp/packages/solo
```

In Windows, use:

```
go get -buildmode=exe github.com/iotaledger/wasp/packages/solo
```

To run Rust/Wasm smart contracts you will also need [Rust](https://www.rust-lang.org/tools/install) and [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/) 
installed. You can use any development environment for Rust and Go. The GoLand
environment with the Rust plugin is a good combination.

You can find example implementations of smart contracts (including source code
and tests) in the Wasp repository, in the
[contracts/rust folder](https://github.com/iotaledger/wasp/tree/master/contracts/rust).
