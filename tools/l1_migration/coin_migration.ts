import { Argument, Transaction } from '@iota/iota-sdk/transactions';
import { gasTypeTag } from './consts';

export namespace CoinMigration {
  export function fromBalance(tx: Transaction, baseTokens: Argument) {
    return tx.moveCall({
      target: '0x2::coin::from_balance',
      typeArguments: [gasTypeTag],
      arguments: [baseTokens],
    });
  }
}
