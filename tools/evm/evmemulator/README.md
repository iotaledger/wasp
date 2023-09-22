# evmemulator

The `evmemulator` tool provides a JSONRPC server with Solo as a backend, allowing
to test Ethereum contracts.

## Example: Uniswap test suite

The following commands will clone and run the Uniswap contract tests against ISC's EVM.

Start the `evmemulator`:

```
evmemulator
```

In another terminal, clone uniswap:

```
git clone https://github.com/Uniswap/uniswap-v3-core.git
yarn install
npx hardhat compile
```

Edit `hardhat.config.ts`, section `networks`:

```
wasp: {
    chainId: 1074,
    url: 'http://localhost:8545',
},
```

Run the test suite:

```
npx hardhat test --network wasp
```
