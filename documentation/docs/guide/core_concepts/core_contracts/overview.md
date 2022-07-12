---
description: There currently are 6 core smart contracts that are always deployed on each  chain, root, _default, accounts, blob, blocklog, and governance.
image: /img/Banner/banner_wasp_core_contracts_overview.png
keywords:
- smart contracts
- core
- initialization
- request handling
- on-chain ledger
- accounts
- data
- receipts
- reference
--- 
# Core Contracts

![Wasp Node Core Contracts Overview](/img/Banner/banner_wasp_core_contracts_overview.png)

There are currently 6 core smart contracts that are always deployed on each
chain. These are responsible for the vital functions of the chain and
provide infrastructure for all other smart contracts:

- [__root__](root.md) - Responsible for the initialization of the chain, maintains registry of deployed contracts.

- [__accounts__](accounts.md): Responsible for the on-chain ledger of accounts (who owns what).

- [__blob__](blob.md): Responsible for the immutable registry of binary objects of arbitrary size. One blob is a collection of named binary chunks of data. For
  example, a blob can be used to store a collections of _wasm_ binaries, needed
  to deploy _WebAssembly_ smart contracts. Each blob in the registry is 
  referenced by its hash which is deterministically calculated from its data.

- [__blocklog__](blocklog.md): Keeps track of the blocks and receipts of requests which were processed by the chain. It also contains all events emitted by smart contracts.

- [__governance__](governance.md): Handles the administrative functions of the chain. For example: rotation of the committee of validators of the chain, fees and other chain-specific configurations.

- [__errors__](errors.md): Keeps a map of error codes to error messages templates. These error codes are used in request receipts.
