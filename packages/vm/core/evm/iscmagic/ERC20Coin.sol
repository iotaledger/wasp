// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCPrivileged.sol";

/**
 * @title ERC20Coin
 * @dev The ERC20 contract for a Iota coin.
 */
contract ERC20Coin {
    using ISCTypes for CoinBalance[];

    uint256 private constant MAX_UINT64 = type(uint64).max;

    string private _iotaCoinType;
    string private _name;
    string private _tickerSymbol;
    uint8 private _decimals;

    /**
     * @dev Emitted when the allowance of a spender for an owner is set.
     * @param tokenOwner The owner of the tokens.
     * @param spender The address allowed to spend the tokens.
     * @param tokens The amount of tokens allowed to be spent.
     */
    event Approval(
        address indexed tokenOwner,
        address indexed spender,
        uint256 tokens
    );

    /**
     * @dev Emitted when tokens are transferred from one address to another.
     * @param from The address tokens are transferred from.
     * @param to The address tokens are transferred to.
     * @param tokens The amount of tokens transferred.
     */
    event Transfer(address indexed from, address indexed to, uint256 tokens);

    /**
     * @dev Returns the name of the native token.
     * @return The name of the token.
     */
    function name() public view returns (string memory) {
        return _name;
    }

    /**
     * @dev Returns the ticker symbol of the native token.
     * @return The ticker symbol of the token.
     */
    function symbol() public view returns (string memory) {
        return _tickerSymbol;
    }

    /**
     * @dev Returns the number of decimals used for the native token.
     * @return The number of decimals.
     */
    function decimals() public view returns (uint8) {
        return _decimals;
    }

    /**
     * @dev Returns the total supply of the native token.
     * @return The total supply of the token.
     */
    function totalSupply() public view virtual returns (uint256) {
        return __iscSandbox.getCoinInfo(_iotaCoinType).totalSupply;
    }

    /**
     * @dev Returns the balance of a token owner.
     * @param tokenOwner The address of the token owner.
     * @return The balance of the token owner.
     */
    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(
            tokenOwner,
            __iscSandbox.getChainID()
        );
        return __iscAccounts.getL2BalanceCoin(_iotaCoinType, ownerAgentID);
    }

    /**
     * @dev Transfers tokens from the sender's address to the receiver's address.
     * @param receiver The address to transfer tokens to.
     * @param numTokens The amount of tokens to transfer.
     * @return true.
     */
    function transfer(
        address receiver,
        uint256 numTokens
    ) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAssets memory assets;
        assets.coins = new CoinBalance[](1);
        assets.coins[0].coinType = _iotaCoinType;
        assets.coins[0].amount = uint64(numTokens);
        __iscPrivileged.moveBetweenAccounts(msg.sender, receiver, assets);
        emit Transfer(msg.sender, receiver, numTokens);
        return true;
    }

    /**
     * @dev Sets the allowance of a spender to spend tokens on behalf of the owner.
     * @param delegate The address allowed to spend the tokens.
     * @param numTokens The amount of tokens allowed to be spent.
     * @return true.
     */
    function approve(
        address delegate,
        uint256 numTokens
    ) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        __iscPrivileged.setAllowanceCoin(
            msg.sender,
            delegate,
            _iotaCoinType,
            uint64(numTokens)
        );
        emit Approval(msg.sender, delegate, numTokens);
        return true;
    }

    /**
     * @dev Returns the amount of tokens that the spender is allowed to spend on behalf of the owner.
     * @param owner The address of the token owner.
     * @param delegate The address of the spender.
     * @return The amount of tokens the spender is allowed to spend.
     */
    function allowance(
        address owner,
        address delegate
    ) public view returns (uint256) {
        ISCAssets memory assets = __iscSandbox.getAllowance(owner, delegate);
        return assets.coins.getCoinAmount(_iotaCoinType);
    }

    /**
     * @dev Transfers tokens from one address to another on behalf of a token owner.
     * @param owner The address from which tokens are transferred.
     * @param buyer The address to which tokens are transferred.
     * @param numTokens The amount of tokens to transfer.
     * @return A boolean indicating whether the transfer was successful or not.
     */
    function transferFrom(
        address owner,
        address buyer,
        uint256 numTokens
    ) public returns (bool) {
        require(numTokens <= MAX_UINT64, "amount is too large");
        ISCAssets memory assets;
        assets.coins = new CoinBalance[](1);
        assets.coins[0].coinType = _iotaCoinType;
        assets.coins[0].amount = uint64(numTokens);
        __iscPrivileged.moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            __iscPrivileged.moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}
