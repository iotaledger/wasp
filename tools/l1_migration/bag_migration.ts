import { Argument, Transaction } from '@iota/iota-sdk/transactions';

export namespace BagMigration {
  export function destroyEmpty(tx: Transaction, bag: Argument) {
    return tx.moveCall({
      target: '0x2::bag::destroy_empty',
      typeArguments: [],
      arguments: [bag],
    });
  }
}
