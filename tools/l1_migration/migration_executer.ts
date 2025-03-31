import { IotaClient } from '@iota/iota-sdk/client';
import { Transaction } from '@iota/iota-sdk/transactions';
import { ISCMove } from './isc';
import { gasTypeTag, nftObjectTypeTag } from './consts';
import { NftMigration } from './nft_migration';
import { AliasMigration } from './alias_migration';
import { CoinMigration } from './coin_migration';
import { BasicMigration } from './basic_migration';
import { BagMigration } from './bag_migration';
import { FoundryMigration } from './foundry_migration';

export async function createMigrationTransaction(client: IotaClient, iscPackageId: string, governorAddress: string, aliasId: string, anchorId: string): Promise<[tx: Transaction, migratedObjects: { nativeToken: {}; NFTs: {}; Foundries: {}; } ]> {
  const aliasOutputId = await AliasMigration.getAliasOutputId(client, aliasId);

  if (!aliasOutputId) {
    throw new Error(`could not get the alias output id: ${aliasId}!`);
  }

  const assets = await AliasMigration.collectNFTsBasicsAndFoundryIDs(client, aliasOutputId);

  /**
   * Here the migration TX gets built.
   *
   * An isc AssetsBag gets created, which will be used to store all assets extracted from the AliasOutput.
   * The assets will be extracted and returned.
   * Then the assets will be pushed into the assets bag. (This is where we deviate from the AliasOutput code example https://docs.iota.org/developer/stardust/claiming/address-unlock-condition)
   */

  let tx = new Transaction();

  let migratedObjects = {
    nativeToken: {},
    NFTs: {},
    Foundries: {}
  }

  const [baseTokens, nativeTokensBag, alias] = AliasMigration.extractAssetsFromAlias(tx, aliasId);
  const aliasBaseCoin = CoinMigration.fromBalance(tx, baseTokens);

  ISCMove.addCoinToAssetsBag(iscPackageId, tx, anchorId, gasTypeTag, aliasBaseCoin);
  BagMigration.destroyEmpty(tx, nativeTokensBag);

  for (let basic of assets.basics) {
    const basicArg = tx.object(basic.objectId);

    const basicOutput = BasicMigration.unlockBasic(tx, alias, basicArg);
    const [baseTokens, nativeTokensBag] = BasicMigration.extractAssetsFromBasicOutput(tx, basicOutput);

    const basicBaseCoin = CoinMigration.fromBalance(tx, baseTokens);
    ISCMove.addCoinToAssetsBag(iscPackageId, tx, anchorId, gasTypeTag, basicBaseCoin);

    const nativeTokens = await BasicMigration.getNativeTokens(client, basic);

    for (const nativeToken of nativeTokens!) {
      const typeArguments = `0x${nativeToken}`;

      const [bag, balance] = BasicMigration.utilitiesExtractBag(tx, typeArguments, nativeTokensBag);

      ISCMove.addBalanceToAssetsBag(iscPackageId, tx, anchorId, typeArguments, balance);
      BagMigration.destroyEmpty(tx, bag);

      migratedObjects.nativeToken[typeArguments] = 1;
    }

    //BagMigration.destroyEmpty(tx, nativeTokensBag);
  }

  for (let nft of assets.nfts) {
    const nftOutputArg = tx.object(nft.objectId);

    const nftOutput = NftMigration.unlockNft(tx, alias, nftOutputArg);
    const [nftBaseTokens, nftNativeTokenBag, nftAsset] = NftMigration.extractAssetsFromNft(tx, nftOutput);
    const nftBaseCoin = CoinMigration.fromBalance(tx, nftBaseTokens);

    BagMigration.destroyEmpty(tx, nftNativeTokenBag);

    ISCMove.addCoinToAssetsBag(iscPackageId, tx, anchorId, gasTypeTag, nftBaseCoin);
    ISCMove.addObjectToAssetsBag(iscPackageId, tx, anchorId, nftObjectTypeTag, nftAsset);

    migratedObjects.NFTs[nft.objectId] = 1;

  }
  
  for (let foundry of assets.foundries) {
    const foundryOutputArg = tx.object(foundry.objectId);
    const foundryTypeArg = foundry.type!.split("<")[1].split(">")[0] || "";
    const foundryOutput = FoundryMigration.unlockFoundry(tx, foundryTypeArg, alias, foundryOutputArg);
    ISCMove.addObjectToAssetsBag(iscPackageId, tx, anchorId, foundry.type!, foundryOutput);

    migratedObjects.Foundries[foundry.type!] = 1;
  }

  tx.transferObjects([anchorId, alias], governorAddress);

  return [tx, migratedObjects];
}
