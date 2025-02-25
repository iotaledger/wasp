import { bcs } from '@iota/iota-sdk/bcs';
import { GetOwnedObjectsParams, IotaClient, IotaObjectResponse, IotaParsedData } from '@iota/iota-sdk/client';
import { paginatedRequest } from './page_reader';
import { Argument, Transaction, TransactionResult } from '@iota/iota-sdk/transactions';
import { ISCMove } from './isc';
import { foundryCapTypeTag, gasTypeTag, nftOutputStructTag, STARDUST_PACKAGE_ID } from './consts';

async function getAliasOutputId(client: IotaClient, aliasId: string) {
  const dfName = {
    type: bcs.TypeTag.serialize({
      vector: {
        u8: true,
      },
    }).parse(),
    value: 'alias',
  };

  const aliasObject = await client.getDynamicFieldObject({ parentId: aliasId, name: dfName });
  if (!aliasObject) {
    throw new Error('Alias object not found');
  }

  // Get the underlying "AliasOutput" ID
  const aliasObjectId = aliasObject.data?.objectId;

  return aliasObjectId;
}

async function collectNFTsAndFoundryIDs(client: IotaClient, aliasOutputId: string) {
  const ownedObjects = await paginatedRequest<IotaObjectResponse, GetOwnedObjectsParams>(x => client.getOwnedObjects(x), {
    owner: aliasOutputId,
    filter: {
      MatchAny: [
        {
          StructType: nftOutputStructTag,
        },
        /* {
          MoveModule: {
            module: 'coin_manager',
            package: '0x2',
          },
        },*/
      ],
    },
    options: {
      showType: true,
      showContent: true,
    },
  });

  // Right now we expect *n* NFTs and a single Foundry capability. In case more foundries roll in, the query still works.
  // But we should make sure we don't accidentially select other than the expected coin_manager types [CoinManagerTreasuryCap].
  const filteredOwnedObjects = ownedObjects.filter(x => x.data?.type == nftOutputStructTag || x.data?.type?.startsWith(foundryCapTypeTag)).map(x => x.data);

  const checkNull = filteredOwnedObjects.filter(x => x == null);
  if (checkNull.length > 0) {
    throw new Error('Failed to get owned objects properly, it contains null item(s)');
  }

  console.log(
    ownedObjects.map(x => {
      const moveObject = x!.data!.content! as IotaParsedData;
      return moveObject['fields']['native_tokens'].fields;
    }),
  );

  return filteredOwnedObjects.filter(x => x != null);
}

async function extractAssetsFromAlias(
  tx: Transaction,
  aliasOutputId: string | TransactionResult,
): Promise<[baseTokenBag: Argument, nativeTokenBag: Argument, alias: Argument]> {
  const ex = tx.moveCall({
    target: `${STARDUST_PACKAGE_ID}::alias_output::extract_assets`,
    typeArguments: [gasTypeTag],
    arguments: [tx.object(aliasOutputId)],
  });

  return [ex[0], ex[1], ex[2]];
}

async function extractAssetsFromNft(tx: Transaction, nftOutput: TransactionResult): Promise<[baseTokenBag: Argument, nativeTokenBag: Argument, nft: Argument]> {
  const ex = tx.moveCall({
    target: `${STARDUST_PACKAGE_ID}::nft_output::extract_assets`,
    typeArguments: [gasTypeTag],
    arguments: [nftOutput],
  });

  return [ex[0], ex[1], ex[2]];
}

export async function consumeAliasOutput(client: IotaClient, iscPackageId: string, aliasId: string) {
  const aliasOutputId = await getAliasOutputId(client, aliasId);

  if (!aliasOutputId) {
    throw new Error(`could not get the alias output id: ${aliasId}!`);
  }
  const assets = await collectNFTsAndFoundryIDs(client, aliasOutputId);
  const assetIds = assets.map(x => x!.objectId);
  
  /**
   * Here the migration TX gets built.
   *
   * An isc AssetsBag gets created, which will be used to store all assets extracted from the AliasOutput.
   * The assets will be extracted and returned.
   * Then the assets will be pushed into the assets bag. (This is where we deviate from the AliasOutput code example https://docs.iota.org/developer/stardust/claiming/address-unlock-condition)
   */

  let tx = new Transaction();

  let assetsBag = ISCMove.newAssetBag(iscPackageId, tx);
  const [baseTokens, nativeTokensBag, alias] = await extractAssetsFromAlias(tx, aliasId);

  const iotaCoin = tx.moveCall({
    target: '0x2::coin::from_balance',
    typeArguments: [gasTypeTag],
    arguments: [baseTokens],
  });

  ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, gasTypeTag, iotaCoin);

  tx.moveCall({
    target: '0x2::bag::destroy_empty',
    typeArguments: [],
    arguments: [nativeTokensBag],
  });

  for (let nft of assetIds) {
    const nftOutputArg = tx.object(nft);

    const nftOutput = tx.moveCall({
      target: `${STARDUST_PACKAGE_ID}::address_unlock_condition::unlock_alias_address_owned_nft`,
      typeArguments: [gasTypeTag],
      arguments: [alias, nftOutputArg],
    });

    const [nftBaseTokens, nftNativeTokenBag, nftAsset] = await extractAssetsFromNft(tx, nftOutput);

    const iotaCoin2 = tx.moveCall({
      target: '0x2::coin::from_balance',
      typeArguments: [gasTypeTag],
      arguments: [nftBaseTokens],
    });

    tx.moveCall({
      target: '0x2::bag::destroy_empty',
      typeArguments: [],
      arguments: [nftNativeTokenBag],
    });

    ISCMove.addCoinToAssetsBag(iscPackageId, tx, assetsBag, gasTypeTag, iotaCoin2);
    ISCMove.addObjectToAssetsBag(iscPackageId, tx, assetsBag, nftAsset);
  }

  tx.setGasBudget(50000000000);

  return tx;
}
