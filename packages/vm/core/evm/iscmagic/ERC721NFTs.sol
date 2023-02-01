// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";
import "@iscmagic/ISCSandbox.sol";
import "@iscmagic/ISCAccounts.sol";
import "@iscmagic/ISCPrivileged.sol";

// The ERC721 contract for the "global" collection of ISC L2 NFTs.
contract ERC721NFTs {
    // is IERC721Metadata, IERC721, IERC165
    using ISCTypes for ISCAgentID;
    using ISCTypes for uint256;

    // Mapping from token ID to approved address
    mapping(uint256 => address) private _tokenApprovals;
    // Mapping from owner to operator approvals
    mapping(address => mapping(address => bool)) private _operatorApprovals;

    event Transfer(
        address indexed from,
        address indexed to,
        uint256 indexed tokenId
    );
    event Approval(
        address indexed owner,
        address indexed approved,
        uint256 indexed tokenId
    );
    event ApprovalForAll(
        address indexed owner,
        address indexed operator,
        bool approved
    );

    function _balanceOf(ISCAgentID memory owner)
        internal
        view
        virtual
        returns (uint256)
    {
        return __iscAccounts.getL2NFTAmount(owner);
    }

    // virtual function meant to be overridden. ERC721NFTs manages all NFTs, regardless of
    // whether they belong to any collection or not.
    function _isManagedByThisContract(ISCNFT memory)
        internal
        view
        virtual
        returns (bool)
    {
        return true;
    }

    function balanceOf(address owner) public view returns (uint256) {
        return _balanceOf(ISCTypes.newEthereumAgentID(owner));
    }

    function ownerOf(uint256 tokenId) public view returns (address) {
        ISCNFT memory nft = __iscSandbox.getNFTData(tokenId.asNFTID());
        require(nft.owner.isEthereum());
        require(_isManagedByThisContract(nft));
        return nft.owner.ethAddress();
    }

    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        bytes memory data
    ) public payable {
        transferFrom(from, to, tokenId);
        require(_checkOnERC721Received(from, to, tokenId, data));
    }

    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public payable {
        safeTransferFrom(from, to, tokenId, "");
    }

    function transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public payable {
        require(_isApprovedOrOwner(msg.sender, tokenId));
        _transferFrom(from, to, tokenId);
    }

    function approve(address approved, uint256 tokenId) public payable {
        address owner = ownerOf(tokenId);
        require(approved != owner);
        require(msg.sender == owner || isApprovedForAll(owner, msg.sender));

        _tokenApprovals[tokenId] = approved;
        emit Approval(owner, approved, tokenId);
    }

    function setApprovalForAll(address operator, bool approved) public {
        require(operator != msg.sender);
        _operatorApprovals[msg.sender][operator] = approved;
        emit ApprovalForAll(msg.sender, operator, approved);
    }

    function getApproved(uint256 tokenId) public view returns (address) {
        return _tokenApprovals[tokenId];
    }

    function isApprovedForAll(address owner, address operator)
        public
        view
        returns (bool)
    {
        return _operatorApprovals[owner][operator];
    }

    function _isApprovedOrOwner(address spender, uint256 tokenId)
        internal
        view
        returns (bool)
    {
        address owner = ownerOf(tokenId);
        return (spender == owner ||
            getApproved(tokenId) == spender ||
            isApprovedForAll(owner, spender));
    }

    function _transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) internal {
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

    // IERC721Metadata

    function name() external view virtual returns (string memory) {
        return "";
    }

    function symbol() external pure returns (string memory) {
        return ""; // not defined in IRC27
    }

    function tokenURI(uint256 tokenId) external view returns (string memory) {
        IRC27NFT memory nft = __iscSandbox.getIRC27NFTData(tokenId.asNFTID());
        require(_isManagedByThisContract(nft.nft));
        return nft.metadata.uri;
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
