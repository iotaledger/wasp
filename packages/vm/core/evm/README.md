# EVM support

This package and subpackages contain the code for the `evm`
core contract, which allows to execute Ethereum VM (EVM) code on top of the
ISC chain, thus adding support for Ethereum smart contracts.

## Installing @iscmagic contracts

The @iscmagic contracts are installable via __NPM__ with

```bash
npm install --save @iotago/iscmagic
```

After installing `@iota/iscmagic` you can use the functions by importing them as you normally would.

```solidity
import "@iota/iscmagic/ISC.sol"
...
...
```

## JSON-RPC

The Wasp node provides a JSON-RPC service at `/chain/<isc-chainid>/evm/jsonrpc`. This will
allow you to connect any standard Ethereum tool, like Metamask. You can check
the Metamask connection parameters for any given ISC chain in the Dashboard.
