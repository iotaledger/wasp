// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCPrivileged.sol";
import "./ERC721NFTs.sol";

/**
 * @title ERC721NFTCollection
 * @dev The ERC721 contract for a L2 collection of ISC NFTs, as defined in IRC27.
 * Implements the ERC721 standard and extends the ERC721NFTs contract.
 * For more information about IRC27, refer to: https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
 */
contract ERC721NFTCollection is ERC721NFTs {
    using ISCTypes for ISCNFT;

    NFTID private _collectionId;
    string private _collectionName; // extracted from the IRC27 metadata

    /**
     * @dev Returns the balance of the specified owner.
     * @param owner The address to query the balance of.
     * @return The balance of the specified owner.
     */
    function _balanceOf(
        ISCAgentID memory owner
    ) internal view virtual override returns (uint256) {
        return __iscAccounts.getL2NFTAmountInCollection(owner, _collectionId);
    }

    /**
     * @dev Checks if the given NFT is managed by this contract.
     * @param nft The NFT to check.
     * @return True if the NFT is managed by this contract, false otherwise.
     */
    function _isManagedByThisContract(
        ISCNFT memory nft
    ) internal view virtual override returns (bool) {
        return nft.isInCollection(_collectionId);
    }

    /**
     * @dev Returns the ID of the collection.
     * @return The ID of the collection.
     */
    function collectionId() external view virtual returns (NFTID) {
        return _collectionId;
    }

    // IERC721Metadata

    /**
     * @dev Returns the name of the collection.
     * @return The name of the collection.
     */
    function name() external view virtual override returns (string memory) {
        return _collectionName;
    }
}
