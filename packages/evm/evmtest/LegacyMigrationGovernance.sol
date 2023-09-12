// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract LegacyMigrationGovernance {
    function burn(L1Address memory chain) public {
        ISCAssets memory assets;
        assets.baseTokens = 1000000; // 1Mi - this assumes the SC is funded externally

        ISCSendMetadata memory metadata;

        metadata.targetContract = ISCHname.wrap(86002025);
        metadata.entrypoint = ISCHname.wrap(2076307377);

        ISCSendOptions memory options;
        ISC.sandbox.send(chain, assets, true, metadata, options);
    }
}
