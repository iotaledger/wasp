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
        _send(claimOwnershipEntrypoint, chain);
    }

    function burn(L1Address memory chain) public {
        _send(burnEntrypoint, chain);
    }

    // _send calls (without parameters) a given entrypoint of the migration contract on a given chain
    function _send(ISCHname entrypoint, L1Address memory chain) private {
        ISCAssets memory assets;
        assets.baseTokens = 1000000; // 1Mi - this assumes the SC is funded externally

        ISCSendMetadata memory metadata;

        metadata.targetContract = migrationContract;
        metadata.entrypoint = entrypoint;

        ISCSendOptions memory options;
        ISC.sandbox.send(chain, assets, true, metadata, options);
    }
}
