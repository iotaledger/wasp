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

[EVM and ISC are inherently very different platforms](compatibility.md).
Some EVM-specific actions (e.g., manipulating Ethereum tokens) are disabled, and EVM contracts can access ISC-specific
functionality through the _ISC Magic Contract_.

The Magic contract is an EVM contract deployed by default on every ISC chain, in the EVM genesis block, at
address `0x1074`.
The implementation of the Magic contract is baked-in in
the [`evm`](../core_concepts/core_contracts/evm.md) [core contract](../core_concepts/core_contracts/overview.md));
i.e. it is not a pure-Solidity contract.

You can access the Magic contract from any Solidity contract by importing
its [interface](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/evm/iscmagic/ISC.sol).

## Examples

### Calling getEntropy()

```solidity
pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract MyEVMContract {
    event EntropyEvent(bytes32 entropy);

    // this will emit a "random" value taken from the ISC entropy value
    function emitEntropy() public {
        bytes32 e = isc.getEntropy();
        emit EntropyEvent(e);
    }
}
```

After `import "@iscmagic/ISC.sol"`, the global variable `isc` points to the Magic contract at `0x1074`, which can be
called like a regular EVM contract.
For example, if you call `isc.getEntropy()` you are calling the `getEntropy` function which, in turn,
calls [ISC Sandbox's](../core_concepts/sandbox.md) `GetEntropy`.


### Calling a native contract
You can call native contracts using `isc.call`

```solidity
pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract MyEVMContract {
    event EntropyEvent(bytes32 entropy);

    function callInccounter() public {
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
        params.items[0] = ISCDictItem("counter", int64Encoded42);
        ISCAllowance memory allowance;
        isc.call(isc.hn("inccounter"), isc.hn("incCounter"), params, allowance);
    }
}
```
`isc.hn` is used to get the `hname` of the incounter countract and the `incCounter` entry point. You can call view entry points using [isc.callView](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/evm/iscmagic/ISC.sol#L67).

The Magic Contract's [interface](https://github.com/iotaledger/wasp/blob/develop/packages/vm/core/evm/iscmagic/ISC.sol)
is well documented, so it doubles as an API reference.

There are some usage examples in
the [ISCTest.sol](https://github.com/iotaledger/wasp/blob/develop/packages/evm/evmtest/ISCTest.sol) contract (used
internally in unit tests).



