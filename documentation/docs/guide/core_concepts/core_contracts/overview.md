# Core Contracts

There are currently 6 core smart contracts that are always deployed on each
chain. These are responsible for the vital functions of the chain and
provide infrastructure for all other smart contracts:

- [__root__](./root.md) - Responsible for the initialization of the chain, maintains registry of deployed contracts.

- [___default__](./default.md) - Any request that cannot be handled by any of the
  other deployed contracts ends up here.

- [__accounts__](./accounts.md) - Responsible for the on-chain ledger of accounts (who owns what).

- [__blob__](./blob.md) - Responsible for the immutable registry of binary objects of arbitrary size. One blob is a collection of named binary chunks of data. For
  example, a blob can be used to store a collections of _wasm_ binaries, needed
  to deploy _WebAssembly_ smart contracts. Each blob in the registry is 
  referenced by its hash which is deterministically calculated from its data.

- [__blocklog__](./blocklog.md) - Keeps track of the blocks and receipts of requests which were processed by the chain. It also contains all events emitted by smart contracts.

- [governance](governance.md) Handles the administrative functions of the chain. For example: rotation of the committee of validators of the chain, fees and other chain-specific configurations.
