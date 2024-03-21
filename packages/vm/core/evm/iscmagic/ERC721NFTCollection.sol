// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCPrivileged.sol";
import "./ERC721NFTs.sol";

// The ERC721 contract for a L2 collection of ISC NFTs, as defined in IRC27:
// https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
contract ERC721NFTCollection is ERC721NFTs {
    using ISCTypes for ISCNFT;

    NFTID private _collectionId;
    string private _collectionName; // extracted from the IRC27 metadata

    function _balanceOf(
        ISCAgentID memory owner
    ) internal view virtual override returns (uint256) {
        return __iscAccounts.getL2NFTAmountInCollection(owner, _collectionId);
    }

    function _isManagedByThisContract(
        ISCNFT memory nft
    ) internal view virtual override returns (bool) {
        return nft.isInCollection(_collectionId);
    }

    function collectionId() external view virtual returns (NFTID) {
        return _collectionId;
    }

    // IERC721Metadata

    function name() external view virtual override returns (string memory) {
        return _collectionName;
    }
}
