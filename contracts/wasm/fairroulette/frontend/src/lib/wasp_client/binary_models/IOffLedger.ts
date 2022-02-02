import type { Buffer } from '../buffer';

export interface OffLedgerArgument {
  key: string;
  value: number;
}

export interface Balance {
  color: Buffer;
  balance: bigint;
}

export interface IOffLedger {
  requestType?: number;
  chainID: Buffer;
  contract: number;
  entrypoint: number;
  arguments: OffLedgerArgument[];
  nonce: bigint;
  balances: Balance[];

  // Public Key and Signature will get set in the Sign function, so no inital set is required
  publicKey?: Buffer;
  signature?: Buffer;
}
