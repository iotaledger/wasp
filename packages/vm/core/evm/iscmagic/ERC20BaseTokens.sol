// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCPrivileged.sol";
import "./ISCAccounts.sol";

// The ERC20 contract for ISC L2 base tokens.
contract ERC20BaseTokens {
    uint256 constant MAX_UINT64 = type(uint64).max;

    event Approval(
        address indexed tokenOwner,
        address indexed spender,
        uint256 tokens
    );
    event Transfer(address indexed from, address indexed to, uint256 tokens);

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
        ISCChainID chainID = __iscSandbox.getChainID();
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(
            tokenOwner,
            chainID
        );
        return __iscAccounts.getL2BalanceBaseTokens(ownerAgentID);
    }

    function transfer(
        address receiver,
        uint256 numTokens
    ) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAssets memory assets;
        assets.baseTokens = uint64(numTokens);
        __iscPrivileged.moveBetweenAccounts(msg.sender, receiver, assets);
        emit Transfer(msg.sender, receiver, numTokens);
        return true;
    }

    // Sets `numTokens` as the allowance of `delegate` over the caller’s tokens.
    //
    // NOTE: Base tokens are represented internally as an uint64.
    //       If numTokens > MAX_UINT64, this call will fail.
    //       Exception: as a special case, numTokens == MAX_UINT256 can be
    //       specified as an "infinite" approval.
    function approve(
        address delegate,
        uint256 numTokens
    ) public returns (bool) {
        __iscPrivileged.setAllowanceBaseTokens(msg.sender, delegate, numTokens);
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    function allowance(
        address owner,
        address delegate
    ) public view returns (uint256) {
        ISCAssets memory assets = __iscSandbox.getAllowance(owner, delegate);
        return assets.baseTokens;
    }

    function transferFrom(
        address owner,
        address buyer,
        uint256 numTokens
    ) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAssets memory assets;
        assets.baseTokens = uint64(numTokens);
        __iscPrivileged.moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            __iscPrivileged.moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}

ERC20BaseTokens constant __erc20BaseTokens = ERC20BaseTokens(
    ISC_ERC20BASETOKENS_ADDRESS
);
