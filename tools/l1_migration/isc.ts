import { Argument, Transaction, TransactionResult } from '@iota/iota-sdk/transactions';
import { nftOutputStructTag } from './consts';


export namespace ISCMove {
  export function newAssetBag(packageId: string, tx: Transaction) {
    const assetsBag = tx.moveCall({
      target: `${packageId}::assets_bag::new`,
      arguments: [],
    });

    return assetsBag;
  }

  export function addCoinToAssetsBag(packageId: string, tx: Transaction, assetsBagId: TransactionResult, coinType: string, coinObject: Argument) {
    const ret = tx.moveCall({
      target: `${packageId}::assets_bag::place_coin`,
      typeArguments: [coinType],
      arguments: [assetsBagId, coinObject],
    });

    return ret;
  }

  export function addObjectToAssetsBag(packageId: string, tx: Transaction, assetsBagId: TransactionResult, assetType: string, object: any) {
    const ret = tx.moveCall({
      target: `${packageId}::assets_bag::place_asset`,
      typeArguments: [assetType],
      arguments: [assetsBagId, object],
    });

    return ret;
  }
}
