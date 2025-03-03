import { IotaClient } from '@iota/iota-sdk/client';
import { Transaction } from '@iota/iota-sdk/transactions';
import { ISCMove } from './isc';
import { gasTypeTag } from './consts';
import { NftMigration } from './nft_migration';
import { AliasMigration } from './alias_migration';
import { CoinMigration } from './coin_migration';
import { BasicMigration } from './basic_migration';
import { BagMigration } from './bag_migration';

export async function executeMigration(client: IotaClient, iscPackageId: string, aliasId: string) {
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

  let assetsBag = ISCMove.newAssetBag(iscPackageId, tx);
  
  const [baseTokens, nativeTokensBag, alias] = AliasMigration.extractAssetsFromAlias(tx, aliasId);
  const aliasBaseCoin = CoinMigration.fromBalance(tx, baseTokens);

  ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, gasTypeTag, aliasBaseCoin);
  BagMigration.destroyEmpty(tx, nativeTokensBag);

  for (let nft of assets.nfts) {
    const nftOutputArg = tx.object(nft.objectId);

    const nftOutput = NftMigration.unlockNft(tx, alias, nftOutputArg);
    const [nftBaseTokens, nftNativeTokenBag, nftAsset] = NftMigration.extractAssetsFromNft(tx, nftOutput);
    const nftBaseCoin = CoinMigration.fromBalance(tx, nftBaseTokens);

    ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, gasTypeTag, nftBaseCoin);
    ISCMove.addObjectToAssetsBag(iscPackageId, tx, assetsBag, nftAsset);

    BagMigration.destroyEmpty(tx, nftNativeTokenBag);
  }

  for (let basic of assets.basics) {
    const basicArg = tx.object(basic.objectId);
    const [baseTokens, nativeTokensBag] = BasicMigration.extractAssetsFromBasicOutput(tx, basicArg);

    const basicBaseCoin = CoinMigration.fromBalance(tx, baseTokens);
    ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, gasTypeTag, basicBaseCoin);

    const nativeTokens = await BasicMigration.getNativeTokens(client, basic);

    for (const nativeToken of nativeTokens!) {
      const typeArguments = `0x${nativeToken}`;

      const [extractedNativeTokenBag, balance] = BasicMigration.utilitiesExtractBag(tx, typeArguments, nativeTokensBag);

      ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, typeArguments, balance);
      BagMigration.destroyEmpty(tx, extractedNativeTokenBag);
    }
  }

  tx.setGasBudget(50000000000);

  return tx;
}
