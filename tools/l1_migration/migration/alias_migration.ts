import { bcs } from '@iota/iota-sdk/bcs';
import { GetOwnedObjectsParams, IotaClient, IotaObjectData, IotaObjectResponse } from '@iota/iota-sdk/client';
import { paginatedRequest } from './page_reader';
import { Argument, Transaction, TransactionResult } from '@iota/iota-sdk/transactions';
import { basicOutputStructTag, foundryCapTypeTag, gasTypeTag, nftOutputStructTag, STARDUST_PACKAGE_ID } from './consts';

export namespace AliasMigration {
  export async function getAliasOutputId(client: IotaClient, aliasId: string) {
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

  export function extractAssetsFromAlias(
    tx: Transaction,
    aliasOutputId: string | TransactionResult,
  ): [baseTokenBag: Argument, nativeTokenBag: Argument, alias: Argument] {
    const ex = tx.moveCall({
      target: `${STARDUST_PACKAGE_ID}::alias_output::extract_assets`,
      typeArguments: [gasTypeTag],
      arguments: [tx.object(aliasOutputId)],
    });

    return [ex[0], ex[1], ex[2]];
  }

  export async function collectNFTsBasicsAndFoundryIDs(client: IotaClient, aliasOutputId: string) {
    const ownedObjects = await paginatedRequest<IotaObjectResponse, GetOwnedObjectsParams>(x => client.getOwnedObjects(x), {
      owner: aliasOutputId,
      filter: {
        MatchAny: [
          {
            StructType: nftOutputStructTag,
          },
          {
            StructType: basicOutputStructTag,
          },
          {
            MoveModule: {
              module: 'coin_manager',
              package: '0x2',
            },
          },
        ],
      },
      options: {
        showType: true,
        showContent: true,
      },
    });

    // check if any object is null or contains an error.
    const invalidObjects = ownedObjects.find(x => x.error || !x.data?.objectId);

    if (invalidObjects) {
      console.log('Invalid objects:', invalidObjects);
      throw new Error('Returned owned objects contain errors');
    }

    const objects: { nfts: Array<IotaObjectData>; basics: Array<IotaObjectData>; foundries: Array<IotaObjectData> } = {
      basics: ownedObjects.filter(x => x.data?.type == basicOutputStructTag).map(x => x.data!),
      foundries: ownedObjects.filter(x => x.data?.type?.startsWith(foundryCapTypeTag)).map(x => x.data!),
      nfts: ownedObjects.filter(x => x.data?.type == nftOutputStructTag).map(x => x.data!),
    };

    return objects;
  }
}
