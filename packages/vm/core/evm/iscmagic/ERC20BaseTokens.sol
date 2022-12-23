// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";
import "@iscmagic/ISCSandbox.sol";
import "@iscmagic/ISCPrivileged.sol";
import "@iscmagic/ISCAccounts.sol";

// The ERC20 contract for ISC L2 base tokens.
contract ERC20BaseTokens {
    uint constant MAX_UINT64 = 1 << 64 - 1;

    event Approval(address indexed tokenOwner, address indexed spender, uint tokens);
    event Transfer(address indexed from, address indexed to, uint tokens);

    function name() public view returns (string memory) {
        return __iscSandbox.getBaseTokenProperties().name;
    }

    function symbol() public view returns (string memory) {
        return __iscSandbox.getBaseTokenProperties().tickerSymbol;
    }

    function decimals() public view returns (uint8) {
        return __iscSandbox.getBaseTokenProperties().decimals;
    }

    function totalSupply() public view returns (uint256) {
        return __iscSandbox.getBaseTokenProperties().totalSupply;
    }

    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(tokenOwner);
        return __iscAccounts.getL2BalanceBaseTokens(ownerAgentID);
    }

    function transfer(address receiver, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        __iscPrivileged.moveBetweenAccounts(msg.sender, receiver, assets);
        emit Transfer(msg.sender, receiver, numTokens);
        return true;
    }

    function approve(address delegate, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        __iscPrivileged.addToAllowance(msg.sender, delegate, assets);
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    function allowance(address owner, address delegate) public view returns (uint) {
        ISCAllowance memory assets = __iscSandbox.getAllowance(owner, delegate);
        return assets.baseTokens;
    }

    function transferFrom(address owner, address buyer, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.baseTokens = uint64(numTokens);
        __iscPrivileged.moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            __iscPrivileged.moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}

ERC20BaseTokens constant __erc20BaseTokens = ERC20BaseTokens(ISC_ERC20BASETOKENS_ADDRESS);
