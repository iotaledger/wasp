# @iota/iscutils

The iscutils package contains various utility methods to simplify the interaction with the IOTA Magic contract. This utility library is designed to be used with the [@iota/iscmagic](https://www.npmjs.com/package/@iota/iscmagic/) npm package.

The Magic contract, an EVM contract, is deployed by default on every ISC chain. It has several methods, accessed via different interfaces like ISCSandbox, ISCAccounts, ISCUtil and more. These can be utilized within any Solidity contract by importing the `@iota/iscmagic` library.

For further information on the Magic contract, check the [Wiki](https://wiki.iota.org/shimmer/smart-contracts/guide/evm/magic/).

## Installing @iota/iscutils contracts

The @iota/iscutils contracts are installable via __NPM__ with

```bash
npm install @iota/iscutils
```

After installing `@iota/iscutils` you can use the functions by importing them as you normally would.

## Utilities

### PRNG

A pseudorandom number generator is available by importing `prng.sol`. Its required seeding process simply depends on a random 32-byte value, and each invocation of the `getRandomNumber()` method will generate an algorithmically determined number, based on the initial seed and numbers generated prior. Its deterministic nature makes it ideal for accomplishing predictable outcomes during testing.

The recommended source for the seed value is attained via the `getEntropy()` method from the ISC sandbox (via the magic contract.) This call yields random bytes drawn from the existing consensus state. Although it is distinctive for each block, it remains the same between immediate calls processed within the same block. Therefore, in order to introduce randomness within a block, this utility comes into play, providing pseudorandom values throughout the block process, buffering the immutable nature of `getEntropy()` within the same block.

#### Example PRNG usage

```solidity
pragma solidity >=0.8.5;

import "@iota/iscmagic/ISC.sol";
import "@iota/iscutils/prng.sol";

contract MyEVMContract {
    using PRNG for PRNG.PRNGState;

    event PseudoRNG(uint256 value);
    event PseudoRNGHash(bytes32 value);
    
    PRNG.PRNGState private prngState;

    function emitValue() public {
        bytes32 e = ISC.sandbox.getEntropy();
        prngState.seed(e);
        bytes32 randomHash = prngState.generateRandomHash();
        emit PseudoRNGHash(randomHash);
        uint256 random = prngState.generateRandomNumber();
        emit PseudoRNG(random);
        uint256 randomRange = prngState.generateRandomNumberInRange(111, 1111111);
        emit PseudoRNG(randomRange);
    }
}
```