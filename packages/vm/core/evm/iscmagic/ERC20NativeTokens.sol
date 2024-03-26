// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCPrivileged.sol";

/**
 * @title ERC20NativeTokens
 * @dev The ERC20 contract native tokens (on-chain foundry).
 */
contract ERC20NativeTokens {
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
     * @dev Returns the foundry serial number of the native token.
     * @return The foundry serial number.
     */
    function foundrySerialNumber() internal view returns (uint32) {
        return __iscSandbox.erc20NativeTokensFoundrySerialNumber(address(this));
    }

    /**
     * @dev Returns the native token ID of the native token.
     * @return The native token ID.
     */
    function nativeTokenID()
        public
        view
        virtual
        returns (NativeTokenID memory)
    {
        return __iscSandbox.getNativeTokenID(foundrySerialNumber());
    }

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
        return
            __iscSandbox
                .getNativeTokenScheme(foundrySerialNumber())
                .maximumSupply;
    }

    /**
     * @dev Returns the balance of a token owner.
     * @param tokenOwner The address of the token owner.
     * @return The balance of the token owner.
     */
    function balanceOf(address tokenOwner) public view returns (uint256) {
        ISCChainID chainID = __iscSandbox.getChainID();
        ISCAgentID memory ownerAgentID = ISCTypes.newEthereumAgentID(
            tokenOwner,
            chainID
        );
        return
            __iscAccounts.getL2BalanceNativeTokens(
                nativeTokenID(),
                ownerAgentID
            );
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
        ISCAssets memory assets;
        assets.nativeTokens = new NativeToken[](1);
        assets.nativeTokens[0].ID = nativeTokenID();
        assets.nativeTokens[0].amount = numTokens;
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
        __iscPrivileged.setAllowanceNativeTokens(
            msg.sender,
            delegate,
            nativeTokenID(),
            numTokens
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
        NativeTokenID memory myID = nativeTokenID();
        for (uint256 i = 0; i < assets.nativeTokens.length; i++) {
            if (bytesEqual(assets.nativeTokens[i].ID.data, myID.data))
                return assets.nativeTokens[i].amount;
        }
        return 0;
    }

    /**
     * @dev Compares two byte arrays for equality.
     * @param a The first byte array.
     * @param b The second byte array.
     * @return A boolean indicating whether the byte arrays are equal or not.
     */
    function bytesEqual(
        bytes memory a,
        bytes memory b
    ) internal pure returns (bool) {
        if (a.length != b.length) return false;
        for (uint256 i = 0; i < a.length; i++) {
            if (a[i] != b[i]) return false;
        }
        return true;
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
        ISCAssets memory assets;
        assets.nativeTokens = new NativeToken[](1);
        assets.nativeTokens[0].ID = nativeTokenID();
        assets.nativeTokens[0].amount = numTokens;
        __iscPrivileged.moveAllowedFunds(owner, msg.sender, assets);
        if (buyer != msg.sender) {
            __iscPrivileged.moveBetweenAccounts(msg.sender, buyer, assets);
        }
        emit Transfer(owner, buyer, numTokens);
        return true;
    }
}
