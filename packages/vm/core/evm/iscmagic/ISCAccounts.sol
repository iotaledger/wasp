// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";

// Functions of the ISC Magic Contract to access the core accounts functionality
interface ISCAccounts {
    // Get the L2 base tokens balance of an account
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) external view returns (uint64);

    // Get the L2 native tokens balance of an account
    function getL2BalanceNativeTokens(NativeTokenID memory id, ISCAgentID memory agentID) external view returns (uint256);
}

ISCAccounts constant __iscAccounts = ISCAccounts(ISC_MAGIC_ADDRESS);
