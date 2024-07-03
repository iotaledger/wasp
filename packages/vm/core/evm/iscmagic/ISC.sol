// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCUtil.sol";
import "./ISCPrivileged.sol";
import "./ERC20BaseTokens.sol";
import "./ERC20NativeTokens.sol";
import "./ERC721NFTs.sol";
import "./ERC721NFTCollection.sol";

/**
 * @title ISC Library
 * @dev This library contains various interfaces and functions related to the IOTA Smart Contracts (ISC) system.
 * It provides access to the ISCSandbox, ISCAccounts, ISCUtil, ERC20BaseTokens, ERC20NativeTokens, ERC721NFTs, and ERC721NFTCollection contracts.
 */
library ISC {
    ISCSandbox constant sandbox = __iscSandbox;

    ISCAccounts constant accounts = __iscAccounts;

    ISCUtil constant util = __iscUtil;

    ERC20BaseTokens constant baseTokens = __erc20BaseTokens;

    /**
     * @notice Get the ERC20NativeTokens contract for the given foundry serial number
     * @param foundrySN The serial number of the foundry
     * @return The ERC20NativeTokens contract corresponding to the given foundry serial number
     */
    function nativeTokens(uint32 foundrySN) internal view returns (ERC20NativeTokens) {
        return ERC20NativeTokens(sandbox.erc20NativeTokensAddress(foundrySN));
    }

    ERC721NFTs constant nfts = __erc721NFTs;

    /**
     * @notice Get the ERC721NFTCollection contract for the given collection
     * @param collectionID The ID of the NFT collection
     * @return The ERC721NFTCollection contract corresponding to the given collection ID
     */
    function erc721NFTCollection(NFTID collectionID) internal view returns (ERC721NFTCollection) {
        return ERC721NFTCollection(sandbox.erc721NFTCollectionAddress(collectionID));
    }

}
