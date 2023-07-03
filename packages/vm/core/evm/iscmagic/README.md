# @iota/iscmagic

The Magic contract is an EVM contract deployed by default on every ISC chain, in the EVM genesis block, at address 0x1074000000000000000000000000000000000000. The implementation of the Magic contract is baked-in in the [evm](https://wiki.iota.org/shimmer/smart-contracts/guide/core_concepts/core_contracts/evm/) [core contract](https://wiki.iota.org/shimmer/smart-contracts/guide/core_concepts/core_contracts/overview/); i.e. it is not a pure-Solidity contract.

The Magic contract has several methods, which are categorized into specialized interfaces: ISCSandbox, ISCAccounts, ISCUtil and so on. You can access these interfaces from any Solidity contract by importing this library.

The Magic contract also provides proxy ERC20 contracts to manipulate ISC base tokens and native tokens on L2.

Read more in the [Wiki](https://wiki.iota.org/shimmer/smart-contracts/guide/evm/magic/).

## Installing @iota/iscmagic contracts

The @iota/iscmagic contracts are installable via __NPM__ with 

```bash
npm install @iota/iscmagic
```

After installing `@iota/iscmagic` you can use the functions by importing them as you normally would.

```ts
pragma solidity >=0.8.5;

import "@iota/iscmagic/ISC.sol";

contract MyEVMContract {
    event EntropyEvent(bytes32 entropy);

    // this will emit a "random" value taken from the ISC entropy value
    function emitEntropy() public {
        bytes32 e = ISC.sandbox.getEntropy();
        emit EntropyEvent(e);
    }
}

```
