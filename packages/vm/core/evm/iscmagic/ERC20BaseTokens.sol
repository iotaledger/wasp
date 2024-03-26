// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCPrivileged.sol";
import "./ISCAccounts.sol";

/**
 * @title ERC20BaseTokens
 * @dev The ERC20 contract directly mapped to the L1 base token.
 */
contract ERC20BaseTokens {
    uint256 private constant MAX_UINT64 = type(uint64).max;

    /**
     * @dev Emitted when the approval of tokens is granted by a token owner to a spender.
     *
     * This event indicates that the token owner has approved the spender to transfer a certain amount of tokens on their behalf.
     *
     * @param tokenOwner The address of the token owner who granted the approval.
     * @param spender The address of the spender who is granted the approval.
     * @param tokens The amount of tokens approved for transfer.
     */
    event Approval(
        address indexed tokenOwner,
        address indexed spender,
        uint256 tokens
    );

    /**
     * @dev Emitted when tokens are transferred from one address to another.
     *
     * This event indicates that a certain amount of tokens has been transferred from one address to another.
     *
     * @param from The address from which the tokens are transferred.
     * @param to The address to which the tokens are transferred.
     * @param tokens The amount of tokens transferred.
     */
    event Transfer(address indexed from, address indexed to, uint256 tokens);

    /**
     * @dev Returns the name of the base token.
     * @return The name of the base token.
     */
    function name() public view returns (string memory) {
        return __iscSandbox.getBaseTokenProperties().name;
    }

    /**
     * @dev Returns the symbol of the base token.
     * @return The symbol of the base token.
     */
    function symbol() public view returns (string memory) {
        return __iscSandbox.getBaseTokenProperties().tickerSymbol;
    }

    /**
     * @dev Returns the number of decimals used by the base token.
     * @return The number of decimals used by the base token.
     */
    function decimals() public view returns (uint8) {
        return __iscSandbox.getBaseTokenProperties().decimals;
    }

    /**
     * @dev Returns the total supply of the base token.
     * @return The total supply of the base token.
     */
    function totalSupply() public view returns (uint256) {
        return __iscSandbox.getBaseTokenProperties().totalSupply;
    }

    /**
     * @dev Returns the balance of the specified token owner.
     * @param tokenOwner The address of the token owner.
     * @return The balance of the token owner.
     */
    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCChainID chainID = __iscSandbox.getChainID();
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(
            tokenOwner,
            chainID
        );
        return __iscAccounts.getL2BalanceBaseTokens(ownerAgentID);
    }

    /**
     * @dev Transfers tokens from the caller's account to the specified receiver.
     * @param receiver The address of the receiver.
     * @param numTokens The number of tokens to transfer.
     * @return true.
     */
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

    /**
     * @dev Sets the allowance of `delegate` over the caller's tokens.
     * @param delegate The address of the delegate.
     * @param numTokens The number of tokens to allow.
     * @return true.
     */
    function approve(
        address delegate,
        uint256 numTokens
    ) public returns (bool) {
        __iscPrivileged.setAllowanceBaseTokens(msg.sender, delegate, numTokens);
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    /**
     * @dev Returns the allowance of the specified owner for the specified delegate.
     * @param owner The address of the owner.
     * @param delegate The address of the delegate.
     * @return The allowance of the owner for the delegate.
     */
    function allowance(
        address owner,
        address delegate
    ) public view returns (uint256) {
        ISCAssets memory assets = __iscSandbox.getAllowance(owner, delegate);
        return assets.baseTokens;
    }

    /**
     * @dev Transfers tokens from the specified owner's account to the specified buyer.
     * @param owner The address of the owner.
     * @param buyer The address of the buyer.
     * @param numTokens The number of tokens to transfer.
     * @return true.
     */
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
