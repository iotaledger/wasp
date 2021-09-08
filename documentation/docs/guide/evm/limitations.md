---
keywords:
- ISCP
- Smart Contracts
- EVM
- Solidity
description: EVM based smart contract limitations
image: /img/logo/WASP_logo_dark.png
---
# EVM limitations within ISCP

The current experimental EVM support for ISCP allows developers to have a sneak preview of EVM based Smart contract solutions on top of IOTA. There are some limitations you should be aware of at this time which we will be adressing in the months to come:

- The current implementation is fully sandboxed and not aware of IOTA or ISCP. It currently can not communicate with non-EVM smart contracts yet nor can it interact with assets outside the EVM sandbox yet.
- You start a EVM chain with a new supply of EVM specific tokens assigned to a single address (the main token on the chain which is used for gas as well, compareable to ETH on the Ethereum network). These new tokens are in no way connected to IOTA or any other token but are specific for that chain for now.
- Because EVM runs inside an ISCP smart contract any fees that need to be paid for that ISCP smart contract have to be taken into account as well while invoking a function on that contract. To support this right now the JSON/RPC gateway simply uses the wallet account connected to it - You need to manually deposit some IOTA to the chain you are using to be able to invoke these functions. We are planning to resolve this at a later phase in a more user friendly way.

Overall these are temporary solutions, the next release of ISCP will see a lot of these improved or resolved.
