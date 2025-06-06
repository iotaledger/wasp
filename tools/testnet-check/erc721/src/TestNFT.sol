pragma solidity ^0.8.0;

import "@openzeppelin/token/ERC721/ERC721.sol";
import "@openzeppelin/access/Ownable.sol";
import "@openzeppelin/token/ERC721/extensions/ERC721Enumerable.sol";

contract TestNFT is ERC721, Ownable {
    uint256 private _nextTokenId;

    constructor(address initialOwner)
    ERC721("IotaEVMSampleNFT", "SSNFT")
    Ownable(initialOwner)
    {}

    function _baseURI() internal pure override returns (string memory) {
        return "https://example.com/nft/";
    }

    function safeMint(address to) public onlyOwner {
        uint256 tokenId = _nextTokenId++;
        _safeMint(to, tokenId);
    }
}