# @iota/iscutils

The iscutils package contains various utility methods to simplify the interaction with the IOTA Magic contract. This utility library is designed to be used with [@iota/iscmagic](https://www.npmjs.com/package/@iota/iscmagic/) npm package.

The Magic contract, an EVM contract, is deployed by default on every ISC chain. It has several methods, accessed via different interfaces like ISCSandbox, ISCAccounts, ISCUtil and more. These can be utilized within any Solidity contract by importing the @iota/iscmagic library.

For further information on the Magic contract, check the [Wiki](https://wiki.iota.org/shimmer/smart-contracts/guide/evm/magic/).

## Installing @iota/iscutils contracts

The @iota/iscutils contracts are installable via __NPM__ with

```bash
npm install @iota/iscutils
```

After installing `@iota/iscutils` you can use the functions by importing them as you normally would.

```solidity
pragma solidity >=0.8.5;

import "@iota/iscmagic/ISC.sol";
import "@iota/iscutils/prng.sol";

contract MyEVMContract {
    using PRNG for PRNG.PRNGState;

    event PseudoRNG(uint256 value);
    
    PRNG.PRNGState private prngState;

    function emitValue() public {
        bytes32 e = ISC.sandbox.getEntropy();
        prngState.seed(e);
        uint256 random = prngState.generateRandomNumber();
        emit PseudoRNG(random);
    }
}

```