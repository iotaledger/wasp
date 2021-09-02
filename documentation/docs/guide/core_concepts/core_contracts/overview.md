# Core Contracts

Deploying a chain automatically means deployment of all core smart contracts on
it. The core contracts are responsible for the vital functions of the chain and
provide infrastructure for all other smart contracts:

- [__root__](./root.md) - Responsible for the initialization of the chain, maintains
  the global parameters, and the registry of deployed contracts. It also handles
  fees and performs other functions.

- [___default__](./default.md) - Any request that cannot be handled by any of the
  other deployed contracts ends up here.

- [__accounts__](./accounts.md) - Responsible for the on-chain ledger of accounts. The
  on-chain accounts contain colored tokens, which are controlled by smart
  contracts and addresses on the UTXO Ledger.

- [__blob__](./blob.md) - Responsible for the immutable registry of binary objects of
  arbitrary size. One blob is a collection of named binary chunks of data. For
  example, a blob can be used to store a collections of _wasm_ binaries, needed
  to deploy _WebAssembly_ smart contracts. Each blob in the registry is 
  referenced by its hash which is deterministically calculated from its data.

- [__blocklog__](./blocklog.md) - Keeps track of the blocks, requests and event that were
  processed by the chain.
