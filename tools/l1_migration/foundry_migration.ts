import { Argument, Transaction } from '@iota/iota-sdk/transactions';
import { STARDUST_PACKAGE_ID } from './consts';

export namespace FoundryMigration {
  export function unlockFoundry(tx: Transaction, typeArg: string, aliasObject: Argument, treasuryCap: Argument) {
    return tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::address_unlock_condition::unlock_alias_address_owned_coinmanager_treasury`,
        typeArguments: [typeArg],
        arguments: [aliasObject, treasuryCap],
      });;
  }
}
