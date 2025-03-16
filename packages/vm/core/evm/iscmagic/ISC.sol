// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import {IotaObjectID} from "./ISCTypes.sol";
import {ISCSandbox, __iscSandbox} from "./ISCSandbox.sol";
import {ISCAccounts, __iscAccounts} from "./ISCAccounts.sol";
import {ISCUtil, __iscUtil} from "./ISCUtil.sol";
import {ERC20Coin} from "./ERC20Coin.sol";
import {ERC721NFTs, __erc721NFTs} from "./ERC721NFTs.sol";
import {ERC721NFTCollection} from "./ERC721NFTCollection.sol";

/**
 * @title ISC Library
 * @dev This library contains various interfaces and functions related to the IOTA Smart Contracts (ISC) system.
 * It provides access to the ISCSandbox, ISCAccounts, ISCUtil,
 * ERC20Coin, ERC721NFTs, and ERC721NFTCollection contracts.
 */
library ISC {
    ISCSandbox constant sandbox = __iscSandbox;

    ISCAccounts constant accounts = __iscAccounts;

    ISCUtil constant util = __iscUtil;

    // Get the ERC20Coin contract for the given foundry serial number
    function erc20Coin(string memory coinType) internal view returns (ERC20Coin) {
        return ERC20Coin(sandbox.ERC20CoinAddress(coinType));
    }

    ERC721NFTs constant nfts = __erc721NFTs;

    // Get the ERC721NFTCollection contract for the given collection
    function erc721NFTCollection(IotaObjectID collectionID) internal view returns (ERC721NFTCollection) {
        return ERC721NFTCollection(sandbox.erc721NFTCollectionAddress(collectionID));
    }

}
