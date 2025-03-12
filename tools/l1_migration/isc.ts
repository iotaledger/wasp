import { bcs } from '@iota/iota-sdk/bcs';
import { Argument, Transaction } from '@iota/iota-sdk/transactions';

export namespace ISCMove {
  export function newAnchor(packageId: string, tx: Transaction, sender: string) {
    const iotaCoinType = '0x0000000000000000000000000000000000000000000000000000000000000002::coin::Coin<0x0000000000000000000000000000000000000000000000000000000000000002::iota::IOTA>'

    const noneOption = tx.moveCall({
      target: '0x1::option::none', // The standard library package
      typeArguments: [iotaCoinType], // Type argument for Coin<IOTA>
      arguments: [],
    });

    // coin: Option<Coin<IOTA>>,
    const ret = tx.moveCall({
      target: `${packageId}::anchor::start_new_chain`,
      arguments: [
        bcs.vector(bcs.u8()).serialize([]), 
        noneOption
      ]});

    tx.transferObjects([ret], sender);

    return ret;
  }

  export function addCoinToAssetsBag(packageId: string, tx: Transaction, anchorId: string, coinType: string, coinObject: Argument) {
    const ret = tx.moveCall({
      target: `${packageId}::anchor::place_coin_for_migration`,
      typeArguments: [coinType],
      arguments: [tx.object(anchorId), coinObject],
    });

    return ret;
  }

  export function addBalanceToAssetsBag(packageId: string, tx: Transaction, anchorId: string, coinType: string, balance: Argument) {
    const ret = tx.moveCall({
      target: `${packageId}::anchor::place_coin_balance_for_migration`,
      typeArguments: [coinType],
      arguments: [tx.object(anchorId), balance],
    });

    return ret;
  }


  export function addObjectToAssetsBag(packageId: string, tx: Transaction, anchorId: string, assetType: string, object: any) {
    const ret = tx.moveCall({
      target: `${packageId}::anchor::place_asset_for_migration`,
      typeArguments: [assetType],
      arguments: [tx.object(anchorId), object],
    });

    return ret;
  }
}
