import { TransactionResult, Argument, Transaction } from '@iota/iota-sdk/transactions';
import { gasTypeTag, STARDUST_PACKAGE_ID } from './consts';

export namespace NftMigration {
  export function extractAssetsFromNft(tx: Transaction, nftOutput: TransactionResult) {
    return tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::nft_output::extract_assets`,
        typeArguments: [gasTypeTag],
        arguments: [nftOutput],
      });
  }

  export function unlockNft(tx: Transaction, aliasObject: Argument, nftObject: Argument) {
    return tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::address_unlock_condition::unlock_alias_address_owned_nft`,
        typeArguments: [gasTypeTag],
        arguments: [aliasObject, nftObject],
      });;
  }
}
