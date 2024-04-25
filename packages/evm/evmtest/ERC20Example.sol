// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

import "@iscmagic/ISC.sol";

contract ERC20Example {
    uint32 private foundrySN;

    function mint(uint256 amount, uint64 storageDeposit) public {
      ISC.accounts.mintNativeTokens(foundrySN, amount, makeAllowanceBaseTokens(storageDeposit));
    }

    function createFoundry(uint256 maxSupply, uint64 storageDeposit) public {
      NativeTokenScheme memory tokenScheme;
      tokenScheme.maximumSupply = maxSupply;
      foundrySN = ISC.accounts.foundryCreateNew(tokenScheme, makeAllowanceBaseTokens(storageDeposit));
    }

    function registerToken(string memory name, string memory symbol, uint8 decimals, uint64 storageDeposit) public {
      ISC.sandbox.registerERC20NativeToken(foundrySN, name, symbol, decimals, makeAllowanceBaseTokens(storageDeposit));
    }

    function createNativeTokenFoundry(string memory tokenName, string memory tokenSymbol, uint8 tokenDecimals, uint256 maxSupply, uint64 storageDeposit) public {
        NativeTokenScheme memory tokenScheme;
        tokenScheme.maximumSupply = maxSupply;
        foundrySN = ISC.accounts.createNativeTokenFoundry(tokenName, tokenSymbol, tokenDecimals, tokenScheme, makeAllowanceBaseTokens(storageDeposit));
    }

    function makeAllowanceBaseTokens(uint64 amount) private pure returns (ISCAssets memory) {
      ISCAssets memory assets;
      assets.baseTokens = amount;
      return assets;
    }
}
