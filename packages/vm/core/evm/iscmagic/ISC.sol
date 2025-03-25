// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCSandbox.sol";
import "./ISCAccounts.sol";
import "./ISCUtil.sol";
import "./ISCPrivileged.sol";
import "./ERC20Coin.sol";

/**
 * @title ISC Library
 * @notice This library contains various interfaces and functions related to the IOTA Smart Contracts (ISC) system.
 * It provides access to the ISCSandbox, ISCAccounts, ISCUtil and ERC20Coin contracts.
 */
library ISC {
    ISCSandbox constant sandbox = __iscSandbox;

    ISCAccounts constant accounts = __iscAccounts;

    ISCUtil constant util = __iscUtil;

    
    /**
     * @notice Retrieves an `ERC20` contract for the specified coin type.
     * @param coinType The type of the coin as a string.
     * @return An `ERC20` contract corresponding to the provided coin type.
     */
    function erc20Coin(string memory coinType) internal view returns (ERC20Coin) {
        return ERC20Coin(sandbox.ERC20CoinAddress(coinType));
    }
}
