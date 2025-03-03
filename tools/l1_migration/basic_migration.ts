import { GetOwnedObjectsParams, IotaClient, IotaObjectData, IotaObjectResponse, IotaParsedData } from '@iota/iota-sdk/client';
import { paginatedRequest } from './page_reader';
import { Argument, Transaction } from '@iota/iota-sdk/transactions';
import { basicOutputStructTag, foundryCapTypeTag, gasTypeTag, nftOutputStructTag, STARDUST_PACKAGE_ID } from './consts';

export namespace BasicMigration {
  export async function getNativeTokens(iotaClient: IotaClient, obj: IotaObjectData) {
    const moveObject = obj.content as IotaParsedData;
    if (moveObject.dataType != 'moveObject') {
      throw new Error('BasicOutput is not a move object');
    }

    // Treat fields as key-value object.
    const fields = moveObject.fields as Record<string, any>;
    const nativeTokensBag = fields['native_tokens'];

    const dfTypeKeys: string[] = [];

    if (nativeTokensBag.fields.size > 0) {
      // Get the dynamic fields owned by the native tokens bag.
      const dynamicFieldPage = await iotaClient.getDynamicFields({
        parentId: nativeTokensBag.fields.id.id,
      });

      // Extract the dynamic fields keys, i.e., the native token type.
      dynamicFieldPage.data.forEach(dynamicField => {
        if (typeof dynamicField.name.value === 'string') {
          dfTypeKeys.push(dynamicField.name.value);
        } else {
          throw new Error('Dynamic field key is not a string');
        }
      });
    }

    return dfTypeKeys!;
  }

  export function extractAssetsFromBasicOutput(tx: Transaction, basicOutput: Argument): [baseTokenBag: Argument, nativeTokenBag: Argument] {
    const ex = tx.moveCall({
      target: `${STARDUST_PACKAGE_ID}::basic_output::extract_assets`,
      typeArguments: [gasTypeTag],
      arguments: [basicOutput],
    });

    return [ex[0], ex[1]];
  }

  export function utilitiesExtractBag(tx: Transaction, typeArg: string, bag: Argument) {
    return tx.moveCall({
      target: `${STARDUST_PACKAGE_ID}::utilities::extract`,
      typeArguments: [typeArg],
      arguments: [bag],
    });
  }
}
