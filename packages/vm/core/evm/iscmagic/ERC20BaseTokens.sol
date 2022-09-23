// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISC.sol";
import "@iscmagic/ISCLib.sol";
import "@iscmagic/ISCPrivileged.sol";

uint constant MAX_UINT64 = 1 << 64 - 1;

// The ERC20 contract for ISC L2 base tokens.
contract ERC20BaseTokens {
    event Approval(address indexed tokenOwner, address indexed spender, uint tokens);
    event Transfer(address indexed from, address indexed to, uint tokens);

    function name() public view returns (string memory) {
        return isc.getBaseTokenProperties().name;
    }

    function symbol() public view returns (string memory) {
        return isc.getBaseTokenProperties().tickerSymbol;
    }

    function decimals() public view returns (uint8) {
        return isc.getBaseTokenProperties().decimals;
    }

    function totalSupply() public view returns (uint256) {
        return isc.getBaseTokenProperties().totalSupply;
    }

    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCAgentID memory ownerAgentID = ISCLib.newEthereumAgentID(tokenOwner);
        return ISCLib.getL2BalanceBaseTokens(ownerAgentID);
    }

    function transfer(address receiver, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        ISCPrivileged(address(isc)).moveBetweenAccounts(msg.sender, receiver, assets);
        emit Transfer(msg.sender, receiver, numTokens);
        return true;
    }

    function approve(address delegate, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        ISCPrivileged(address(isc)).addToAllowance(msg.sender, delegate, assets);
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    function allowance(address owner, address delegate) public view returns (uint) {
        ISCAllowance memory assets = isc.getAllowance(owner, delegate);
        return assets.baseTokens;
    }

    function transferFrom(address owner, address buyer, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        ISCPrivileged(address(isc)).moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            ISCPrivileged(address(isc)).moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}

// Every ISC chain is initialized with an instance of the ERC20BaseTokens contract at address 0x1075
ERC20BaseTokens constant erc20BaseTokens = ERC20BaseTokens(address(uint160(ISC_ADDRESS + 1)));
