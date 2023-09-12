![Wasp logo](https://github.com/iotaledger/iota-wiki/blob/main/static/img/logo/WASP_logo_dark.png)

# Legacy-Migration Wasp

For the regular wasp node software go to <https://github.com/iotaledger/wasp> .

<!-- TODO add link -->
This version of the wasp software was made with the objective of allowing the [migration of legacy funds](.) from the pre-chysalis network.

Contains a simple [contract](./packages/legacymigration/interface.go) that can be called to release funds given a valid signature of an unmigrated bundle.

This repo also contains a [snapshot](./packages/legacymigration/migratable.csv) of the old network containing all the unmigrated bundles

Everytime a migration is successful, the funds are released on L1 to the target address and an [event](./packages/legacymigration/impl.go:102) is published, making the entire process auditable.

<!-- TODO add link -->
At any point, the [governance contract of the EVM Chain](.) can vote and decide to burn the unmigrated tokens.

## Instructions

<!-- TODO  create node-docker-setup for legacy-migration and add link-->
All committee participants must use the [node-docker-setup with the legacy-migration wasp software](.).

The deployer must use the [wasp-cli compiled from this branch](https://github.com/iotaledger/wasp/tree/legacy-migration).

- deploy the chain

```shell
wasp-cli chain deploy --chain=migration-chain --migration-admin=0x0000000000000000000000000000000000000000@IOTA_EVM_CHAIN_ID
```

then activate the chain on each committee node

- funds must be deposited to the contract (substitute `MIGRATION_CHAIN_ID` and `AMOUNT`):

```shell
wasp-cli chain post-request -s accounts transferAllowanceTo string a agentid 05204969@MIGRATION_CHAIN_ID --transfer=base:AMOUNT --allowance=base:AMOUNT
```

- call a view to confirm everything is as expected

```shell
wasp-cli chain call-view legacymigration getTotalBalance | tee >(wasp-cli decode string total uint64) >(wasp-cli decode string balance uint64)
```

- (optional) relinquish control of the chain governance/UTXO

```shell
wasp-cli chain change-gov-controller <NEXT_OWNER_ADDRESS>
```

By setting the governance controller of the chain UTXO to the chain itself, it means only the chain itself can rotate the committee of validators. It's also possible to set the gov-controller to some unusable address, like the Zero address (`0x000...`), but that means the committee cannot ever be changed.

If the chain gov controller is set to the chain itself, committee changes and other administrative tasks can be made by the "chain admin" (including changing who the chain admin is), by interacting with the [Governance Contract](https://wiki.iota.org/wasp-evm/reference/core-contracts/governance/)

---

## Testing Instructions (Localhost)

(this assumes `docker` and `docker compose` are installed)

- edit `packages/legacymigration/migratable.csv` to include the test cases
- `make install` - this will make sure the software build, and will make wasp-cli available
- `cd tools/local-setup`
- create docker volumes for wasp / hornet:

```shell
docker volume create --name hornet-nest-db
docker volume create --name wasp-db
```

- build and start the test node `bash build_container.sh && docker compose up`
- on another directory of your choice, setup wasp-cli and deploy a chain:

```shell
wasp-cli init
wasp-cli set l1.apiaddress http://localhost:14265
wasp-cli set l1.faucetaddress http://localhost:8091
wasp-cli wasp add my-node http://localhost:9090
wasp-cli request-funds
wasp-cli chain deploy --chain=test-migration-chain --migration-admin=0x0000000000000000000000000000000000000000@tst1pq3e0awlpdsvk7uh2dcaz86mp63nflypu6fyprwxvfherr8lksktwhdwp9j
```

(`migration-admin` is just a dummy address, should be an entity of IotaEVM in the real network)

- send funds to the migration contract (it must be at least the total amount specified in `migratable.csv` (sum of all migratable balances)). In the following command,`CHAIN_ID` must be substituted by the ChainID of the newly deployed chain, also `1000000` is just a placeholder value that should be substituted with the correct amount to deposit.

```shell
wasp-cli chain post-request -s accounts transferAllowanceTo string a agentid 05204969@CHAIN_ID --transfer=base:1000000 --allowance=base:1000000
```

- ensure everything looks good by calling the view entrypoints of the legacy-migration contract:

```shell
wasp-cli chain call-view legacymigration getTotalBalance | tee >(wasp-cli decode string total uint64) >(wasp-cli decode string balance uint64)

```

`balance` must be >= `total`

optionally, you can check how many tokens are migratable for a given address with:

```shell
wasp-cli chain call-view legacymigration getMigratableBalance string a bytes 0x...
```

where `0x...` is the binary format of the migratable address in t5b1 encoding (`t5b1.EncodeTrytes`)

- off-ledger requests can be sent to `http://localhost/wasp/api/v1/requests/offledger` with the following body: `{Request: "Hex string"}`
