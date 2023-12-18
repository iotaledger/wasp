// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract LegacyMigrationGovernance {
    ISCHname public constant migrationContract = ISCHname.wrap(86002025);
    ISCHname public constant burnEntrypoint = ISCHname.wrap(2076307377);
    ISCHname public constant claimOwnershipEntrypoint =
        ISCHname.wrap(3115287444);

    function claimOwnership(L1Address memory chain) public {
        ISCDict memory params = ISCDict(new ISCDictItem[](0));
        _send(claimOwnershipEntrypoint, chain, params);
    }

    function withdraw(
        L1Address memory chain,
        L1Address memory targetAddr
    ) public {
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        params.items[0] = ISCDictItem("targetAddress", targetAddr.data);
        _send(burnEntrypoint, chain, params);
    }

    // _send calls a given entrypoint of the migration contract on a given chain
    function _send(
        ISCHname entrypoint,
        L1Address memory chain,
        ISCDict memory params
    ) private {
        ISCAssets memory assets;
        assets.baseTokens = 1000000; // 1Mi - this assumes the current contract is funded externally

        ISCSendMetadata memory metadata;

        metadata.targetContract = migrationContract;
        metadata.entrypoint = entrypoint;
        metadata.params = params;

        ISCSendOptions memory options;
        ISC.sandbox.send(chain, assets, true, metadata, options);
    }
}
