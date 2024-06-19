// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCPrivileged.sol";

/**
 * @title ERC721NFTs
 * @dev This contract represents the ERC721 contract for the "global" collection of native NFTs on the chains L1 account.
 */
contract ERC721NFTs {
    // is IERC721Metadata, IERC721, IERC165
    using ISCTypes for ISCAgentID;
    using ISCTypes for uint256;

    // Mapping from token ID to approved address
    mapping(uint256 => address) private _tokenApprovals;
    // Mapping from owner to operator approvals
    mapping(address => mapping(address => bool)) private _operatorApprovals;

    /**
     * @dev Emitted when a token is transferred from one address to another.
     *
     * @param from The address transferring the token.
     * @param to The address receiving the token.
     * @param tokenId The ID of the token being transferred.
     */
    event Transfer(
        address indexed from,
        address indexed to,
        uint256 indexed tokenId
    );

    /**
     * @dev Emitted when the approval of a token is changed or reaffirmed.
     *
     * @param owner The owner of the token.
     * @param approved The new approved address.
     * @param tokenId The ID of the token.
     */
    event Approval(
        address indexed owner,
        address indexed approved,
        uint256 indexed tokenId
    );

    /**
     * @dev Emitted when operator gets the allowance from owner.
     *
     * @param owner The owner of the token.
     * @param operator The operator to get the approval.
     * @param approved True if the operator got approval, false if not.
     */
    event ApprovalForAll(
        address indexed owner,
        address indexed operator,
        bool approved
    );

    function _balanceOf(
        ISCAgentID memory owner
    ) internal view virtual returns (uint256) {
        return __iscAccounts.getL2NFTAmount(owner);
    }

    // virtual function meant to be overridden. ERC721NFTs manages all NFTs, regardless of
    // whether they belong to any collection or not.
    function _isManagedByThisContract(
        ISCNFT memory
    ) internal view virtual returns (bool) {
        return true;
    }

    /**
     * @dev Returns the number of tokens owned by a specific address.
     * @param owner The address to query the balance of.
     * @return The balance of the specified address.
     */
    function balanceOf(address owner) public view returns (uint256) {
        ISCChainID chainID = __iscSandbox.getChainID();
        return _balanceOf(ISCTypes.newEthereumAgentID(owner, chainID));
    }

    /**
     * @dev Returns the owner of the specified token ID.
     * @param tokenId The ID of the token to query the owner for.
     * @return The address of the owner of the token.
     */
    function ownerOf(uint256 tokenId) public view returns (address) {
        try __iscSandbox.getNFTData(tokenId.asNFTID()) returns (
            ISCNFT memory nft
        ) {
            require(nft.owner.isEthereum());
            require(_isManagedByThisContract(nft));
            return nft.owner.ethAddress();
        } catch {
            revert("ERC721NonexistentToken");
        }
    }

    function _requireNftExists(uint256 tokenId) internal view {
        ownerOf(tokenId); // ownderOf will revert if the NFT does not exist
    }

    /**
     * @dev Safely transfers an ERC721 token from one address to another.
     *
     * Emits a `Transfer` event.
     *
     * Requirements:
     * - `from` cannot be the zero address.
     * - `to` cannot be the zero address.
     * - The token must exist and be owned by `from`.
     * - If `to` is a smart contract, it must implement the `onERC721Received` function and return the magic value.
     *
     * @param from The address to transfer the token from.
     * @param to The address to transfer the token to.
     * @param tokenId The ID of the token to be transferred.
     * @param data Additional data with no specified format, to be passed to the `onERC721Received` function if `to` is a smart contract.
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        bytes memory data
    ) public payable {
        transferFrom(from, to, tokenId);
        require(_checkOnERC721Received(from, to, tokenId, data));
    }

    /**
     * @dev Safely transfers an ERC721 token from one address to another.
     *
     * Emits a `Transfer` event.
     *
     * Requirements:
     * - `from` cannot be the zero address.
     * - `to` cannot be the zero address.
     * - The caller must own the token or be approved for it.
     *
     * @param from The address to transfer the token from.
     * @param to The address to transfer the token to.
     * @param tokenId The ID of the token to be transferred.
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public payable {
        safeTransferFrom(from, to, tokenId, "");
    }

    /**
     * @dev Transfers an ERC721 token from one address to another.
     * Emits a {Transfer} event.
     *
     * Requirements:
     * - The caller must be approved or the owner of the token.
     *
     * @param from The address to transfer the token from.
     * @param to The address to transfer the token to.
     * @param tokenId The ID of the token to be transferred.
     */
    function transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public payable {
        require(_isApprovedOrOwner(msg.sender, tokenId));
        _transferFrom(from, to, tokenId);
    }

    /**
     * @dev Approves another address to transfer the ownership of a specific token.
     * @param approved The address to be approved for token transfer.
     * @param tokenId The ID of the token to be approved for transfer.
     * @notice Only the owner of the token or an approved operator can call this function.
     */
    function approve(address approved, uint256 tokenId) public payable {
        address owner = ownerOf(tokenId);
        require(approved != owner);
        require(msg.sender == owner || isApprovedForAll(owner, msg.sender));

        _tokenApprovals[tokenId] = approved;
        emit Approval(owner, approved, tokenId);
    }

    /**
     * @dev Sets or revokes approval for the given operator to manage all of the caller's tokens.
     * @param operator The address of the operator to set approval for.
     * @param approved A boolean indicating whether to approve or revoke the operator's approval.
     */
    function setApprovalForAll(address operator, bool approved) public {
        require(operator != msg.sender);
        _operatorApprovals[msg.sender][operator] = approved;
        emit ApprovalForAll(msg.sender, operator, approved);
    }

    /**
     * @dev Returns the address that has been approved to transfer the ownership of the specified token.
     * @param tokenId The ID of the token.
     * @return The address approved to transfer the ownership of the token.
     */
    function getApproved(uint256 tokenId) public view returns (address) {
        _requireNftExists(tokenId);
        return _tokenApprovals[tokenId];
    }

    /**
     * @dev Checks if an operator is approved to manage all of the owner's tokens.
     * @param owner The address of the token owner.
     * @param operator The address of the operator.
     * @return A boolean value indicating whether the operator is approved for all tokens of the owner.
     */
    function isApprovedForAll(
        address owner,
        address operator
    ) public view returns (bool) {
        return _operatorApprovals[owner][operator];
    }

    function _isApprovedOrOwner(
        address spender,
        uint256 tokenId
    ) internal view returns (bool) {
        address owner = ownerOf(tokenId);
        return (spender == owner ||
            getApproved(tokenId) == spender ||
            isApprovedForAll(owner, spender));
    }

    function _transferFrom(address from, address to, uint256 tokenId) internal {
        require(ownerOf(tokenId) == from);
        require(to != address(0));
        _clearApproval(tokenId);

        ISCAssets memory allowance;
        allowance.nfts = new NFTID[](1);
        allowance.nfts[0] = tokenId.asNFTID();

        __iscPrivileged.moveBetweenAccounts(from, to, allowance);

        emit Transfer(from, to, tokenId);
    }

    function _clearApproval(uint256 tokenId) private {
        if (_tokenApprovals[tokenId] != address(0)) {
            _tokenApprovals[tokenId] = address(0);
        }
    }

    // ERC165

    bytes4 private constant _INTERFACE_ID_ERC721METADATA = 0x5b5e139f;
    bytes4 private constant _INTERFACE_ID_ERC721 = 0x80ac58cd;
    bytes4 private constant _INTERFACE_ID_ERC165 = 0x01ffc9a7;

    /**
     * @dev Checks if a contract supports a given interface.
     * @param interfaceID The interface identifier.
     * @return A boolean value indicating whether the contract supports the interface.
     */
    function supportsInterface(bytes4 interfaceID) public pure returns (bool) {
        return
            interfaceID == _INTERFACE_ID_ERC165 ||
            interfaceID == _INTERFACE_ID_ERC721 ||
            interfaceID == _INTERFACE_ID_ERC721METADATA;
    }

    bytes4 private constant _ERC721_RECEIVED = 0x150b7a02;

    function _checkOnERC721Received(
        address from,
        address to,
        uint256 tokenId,
        bytes memory data
    ) internal returns (bool) {
        if (!_isContract(to)) {
            return true;
        }

        bytes4 retval = IERC721Receiver(to).onERC721Received(
            msg.sender,
            from,
            tokenId,
            data
        );
        return (retval == _ERC721_RECEIVED);
    }

    function _isContract(address account) internal view returns (bool) {
        uint256 size;
        assembly {
            size := extcodesize(account)
        }
        return size > 0;
    }

    function name() external view virtual returns (string memory) {
        return "L1 NFTs";
    }

    function symbol() external pure returns (string memory) {
        return "CollectionL1";
    }

    // IERC721Metadata
    function tokenURI(uint256 tokenId) external view returns (string memory) {
        _requireNftExists(tokenId);
        string memory uri = __iscSandbox.getIRC27TokenURI(tokenId.asNFTID());
        return uri;
    }
}

ERC721NFTs constant __erc721NFTs = ERC721NFTs(ISC_ERC721_ADDRESS);

interface IERC721Receiver {
    function onERC721Received(
        address _operator,
        address _from,
        uint256 _tokenId,
        bytes memory _data
    ) external returns (bytes4);
}
