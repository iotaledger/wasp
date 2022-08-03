---
description: The ISC Magic Contract allows EVM contracts to access ISC functionality.
image: /img/logo/WASP_logo_dark.png
keywords:
- configure
- using
- EVM
- magic
- Ethereum
- Solidity
- metamask
- JSON
- RPC
---
# The ISC Magic Contract

As previously [noted](compatibility.md), EVM and ISC are inherently very different platforms.
As such, some EVM-specific actions (e.g. manipulating Ethereum tokens) is disabled, and ISC-specific functionality can be accessed by EVM contracts through the _ISC Magic Contract_.

The Magic contract is an EVM contract that is deployed by default on every ISC chain, in the EVM genesis block, at address `0x1074`.
The implementation of the Magic contract is baked-in in the `evm` core contract; i.e. it is not a pure-Solidity contract.

The Magic contract can be accessed from any Solidity contract by importing its [interface](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/evm/iscmagic/ISC.sol). For example:

```solidity
pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract MyEVMContract {
    event EntropyEvent(bytes32 entropy);

    // this will emit a "random" value, taken from the ISC entropy value
    function emitEntropy() public {
        bytes32 e = isc.getEntropy();
        emit EntropyEvent(e);
    }
}
```

After `import "@iscmagic/ISC.sol"`, the global variable `isc` points to the
Magic contract at `0x1074`, which can be called like a regular EVM contract.
For example, with `isc.getEntropy()` we are calling the `getEntropy` function which, in turn, calls [ISC Sandbox's](../core_concepts/sandbox.md) `GetEntropy`.

The Magic Contract's [interface](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/evm/iscmagic/ISC.sol) is well documented, so it doubles as an API reference.

There are some usage examples in the [ISCTest.sol](https://github.com/iotaledger/wasp/blob/develop/packages/evm/evmtest/ISCTest.sol) contract (used internally in unit tests).
