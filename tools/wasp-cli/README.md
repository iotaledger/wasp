# Wasp Client tool

`wasp-cli` is a command line tool for interacting with Wasp and its smart
contracts.

**Note:** `wasp-cli` is in its early stages, only suitable for testing
purposes.

Flags common to all subcommands:

* `-w`: Wait for requests to complete before returning
* `-v`: Be verbose
* `-c <filename>`: Use given config file. Default: `wasp-cli.json`

## Configuring wasp & goshimmer nodes

`wasp-cli` expects a Goshimmer node and a cluster of wasp nodes to be
accessible.

Default values for node locations:

```
goshimmer.api: 127.0.0.1:8080

wasp.0.api: 127.0.0.1:9090
wasp.0.peering: 127.0.0.1:4000

wasp.1.api: 127.0.0.1:9091
wasp.1.peering: 127.0.0.1:4001

...etc
```

All commands that need to contact a single wasp node use `wasp.0`.

To change the configuration (saved
in `wasp-cli.json`): `wasp-cli set <name> <value>`

Example: `wasp-cli set wasp.1.api wasp1.example.com:9091`

## IOTA wallet

`wasp-cli` provides the following commands for manipulating an IOTA wallet:

* Create a new wallet seed (creates `wasp-cli.json` which stores the
  seed): `wasp-cli init`

  **Note:** `wasp-cli` is alpha phase. The seed is currently being stored in a
  plain text file, which is NOT secure; do not use this seed to store funds in
  the mainnet!

* Show private key + public key + account address for index 0 (index optional,
  default 0): `wasp-cli address [-i index]`

* Query Goshimmer for account balance: `wasp-cli balance [-i index]`

* Use Testnet Faucet to transfer some funds into the wallet address at index
  n: `wasp-cli request-funds [-i index]`

## Working with chains

* List the currently deployed chains: `wasp-cli chain list`

* Deploy a
  chain: `wasp-cli chain deploy --chain=<alias> --nodes=<node indices> --quorum=<T>`

Example:

```
wasp-cli chain deploy --chain=mychain --nodes='0,1,2,3' --quorum=3 --description="My chain"
```

* Set the chain alias for future commands (automatically done after deploying a
  chain): `wasp-cli set chain <alias>`

* List all contracts in the chain: `wasp-cli chain list-contracts`

* List all accounts in the chain: `wasp-cli chain list-accounts`

* Display the in-chain balance of an agentid: `wasp-cli chain balance <agentid>`

## Working with contracts

* Deploy a
  contract: `wasp-cli chain deploy-contract <vmtype> <sc-name> <description> <wasm-file>`

Example: `wasp-cli chain deploy-contract wasmtime inccounter "inccounter SC" contracts/wasm/inccounter_bg.wasm`

* Post a request: `wasp-cli chain post-request <sc-name> <func-name> [args...]`

Example: `wasp-cli chain post-request inccounter increment`

* Call a view: `wasp-cli chain call-view <sc-name> <func-name> [args...]`

Example: `wasp-cli chain call-view inccounter incrementViewCounter`

This command returns a json-encoded representation of the return value, but it
is currently not human-readable (since keys and values are uninterpreted byte
arrays).

* Decode view return value given a schema: `wasp-cli decode <schema>`

Example: `wasp-cli chain call-view inccounter incrementViewCounter | wasp-cli decode string counter int`
