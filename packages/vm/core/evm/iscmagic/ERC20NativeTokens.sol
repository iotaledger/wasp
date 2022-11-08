// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";
import "@iscmagic/ISCSandbox.sol";
import "@iscmagic/ISCAccounts.sol";
import "@iscmagic/ISCPrivileged.sol";

// The ERC20 contract for ISC L2 native tokens.
contract ERC20NativeTokens {
    uint constant MAX_UINT64 = 1 << 64 - 1;

    string _name;
    string _tickerSymbol;
    uint8 _decimals;

    event Approval(address indexed tokenOwner, address indexed spender, uint tokens);
    event Transfer(address indexed from, address indexed to, uint tokens);

    function foundrySerialNumber() public view returns (uint32) {
        return __iscSandbox.erc20NativeTokensFoundrySerialNumber(address(this));
    }

    function nativeTokenID() public view returns (NativeTokenID memory) {
        return __iscSandbox.getNativeTokenID(foundrySerialNumber());
    }

    function name() public view returns (string memory) {
        return _name;
    }

    function symbol() public view returns (string memory) {
        return _tickerSymbol;
    }

    function decimals() public view returns (uint8) {
        return _decimals;
    }

    function totalSupply() public view returns (uint256) {
        return __iscSandbox.getNativeTokenScheme(foundrySerialNumber()).maximumSupply;
    }

    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(tokenOwner);
        return __iscAccounts.getL2BalanceNativeTokens(nativeTokenID(), ownerAgentID);
    }

    function transfer(address receiver, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.tokens = new NativeToken[](1);
        assets.tokens[0].ID = nativeTokenID();
        assets.tokens[0].amount = numTokens;
        __iscPrivileged.moveBetweenAccounts(msg.sender, receiver, assets);
        emit Transfer(msg.sender, receiver, numTokens);
        return true;
    }

    function approve(address delegate, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.tokens = new NativeToken[](1);
        assets.tokens[0].ID = nativeTokenID();
        assets.tokens[0].amount = numTokens;
        __iscPrivileged.addToAllowance(msg.sender, delegate, assets);
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    function allowance(address owner, address delegate) public view returns (uint) {
        ISCAllowance memory assets = __iscSandbox.getAllowance(owner, delegate);
        NativeTokenID memory myID = nativeTokenID();
        for (uint i = 0; i < assets.tokens.length; i++) {
            if (memcmp(assets.tokens[i].ID.data, myID.data))
                return assets.tokens[i].amount;
        }
        return 0;
    }

    function memcmp(bytes memory a, bytes memory b) internal pure returns (bool) {
        return (a.length == b.length) && (keccak256(a) == keccak256(b));
    }

    function transferFrom(address owner, address buyer, uint256 numTokens) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAllowance memory assets;
        assets.tokens = new NativeToken[](1);
        assets.tokens[0].ID = nativeTokenID();
        assets.tokens[0].amount = numTokens;
        __iscPrivileged.moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            __iscPrivileged.moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}
