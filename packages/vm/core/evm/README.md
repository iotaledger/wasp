# EVM support

This package and subpackages contain the code for the `evm`
core contract, which allows to execute Ethereum VM (EVM) code on top of the
ISC chain, thus adding support for Ethereum smart contracts.

## JSON-RPC

The Wasp node provides a JSON-RPC service at `/chain/<isc-chainid>/evm/jsonrpc`. This will
allow you to connect any standard Ethereum tool, like Metamask. You can check
the Metamask connection parameters for any given ISC chain in the Dashboard.

## Complete example using `wasp-cluster`

1. Start a test cluster:

    ```
    wasp-cluster start -d
    ```

2. In a different terminal, initialize a private key and request some iotas from the faucet:

    ```
    wasp-cli init
    wasp-cli request-funds
    ```

3. Deploy an ISC chain with an arbitrary Ethereum chain ID (which should be
   different from any standard Ethereum chain IDs -- see
   https://chainlist.org):

    ```
    wasp-cli chain deploy --chain=mychain --evm-chainid 1234
    ```

4. Send some iotas from your L1 account to any Ethereum account on L2 (e.g. to cover for gas fees):

    ```
    wasp-cli chain deposit 0xa1b2c3d4... iota:1000000
    ```

5. Visit the Wasp dashboard (`<URL>/wasp/dashboard` when using `node-docker-setup`), go to `Chains`, then to
   your ISC chain, scroll down and you will find the EVM section with the
   JSON-RPC URL for Metamask or any other Ethereum tool.

You can now deploy an EVM contract like you would on Ethereum.

For more information check out the [docs](https://wiki.iota.org/smart-contracts/guide/evm/introduction).
