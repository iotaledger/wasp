// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCSandbox.sol";
import "@iscmagic/ISCAccounts.sol";
import "@iscmagic/ISCUtil.sol";
import "@iscmagic/ISCPrivileged.sol";
import "@iscmagic/ERC20BaseTokens.sol";
import "@iscmagic/ERC20NativeTokens.sol";

library ISC {
    ISCSandbox constant sandbox = __iscSandbox;

    ISCAccounts constant accounts = __iscAccounts;

    ISCUtil constant util = __iscUtil;

    ERC20BaseTokens constant erc20BaseTokens = __erc20BaseTokens;

    // Get the ERC20NativeTokens contract for the given foundry serial number
    function erc20NativeTokens(uint32 foundrySN) internal view returns (ERC20NativeTokens) {
        return ERC20NativeTokens(sandbox.erc20NativeTokensAddress(foundrySN));
    }
}
