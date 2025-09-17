# evmemulator

The `evmemulator` tool provides a JSONRPC server with Solo as a backend, allowing
to test Ethereum contracts.

## Hive vs Non-Hive Modes

- `--hive=false` (default): Uses IOTA native behavior with chain ID `1074`, default EVM gas policy, and initializes the full set of development Ethereum accounts with L2 funds.
- `--hive=true`: Uses Hive behavior with chain ID `1`, cheaper gas (for Hive tests), prefunds all accounts from the provided genesis file into L2, and initializes a single development Ethereum account.

Note: When a `--genesis <path>` is provided, both modes load it. Only Hive mode performs the additional prefunding loop for all genesis allocations.

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
