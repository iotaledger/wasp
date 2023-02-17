// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";
import "@iscmagic/ISCSandbox.sol";
import "@iscmagic/ISCAccounts.sol";
import "@iscmagic/ISCPrivileged.sol";
import "@iscmagic/ERC721NFTs.sol";

// The ERC721 contract for a L2 collection of ISC NFTs, as defined in IRC27:
// https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
contract ERC721NFTCollection is ERC721NFTs {
    using ISCTypes for ISCNFT;

    NFTID _collectionId;
    string _collectionName; // extracted from the IRC27 metadata

    function _balanceOf(ISCAgentID memory owner) internal virtual override view returns (uint256) {
        return __iscAccounts.getL2NFTAmountInCollection(owner, _collectionId);
    }

    function _isManagedByThisContract(ISCNFT memory nft) internal virtual override view returns (bool) {
        return nft.isInCollection(_collectionId);
    }

    // IERC721Metadata

    function name() external virtual override view returns (string memory) {
        return _collectionName;
    }
}
