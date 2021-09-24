import type { BasicClient } from './basic_client';
import { BasicWallet } from './basic_wallet';
import { Faucet, IFaucetRequest } from './binary_models';
import type { IOnLedger } from './binary_models/IOnLedger';
import { OnLedger } from './binary_models/on_ledger';
import { Buffer } from './buffer';
import { Base58 } from './crypto';
import type { IKeyPair } from './models';
import type { ITransaction } from './models/ITransaction';
import type { IUnlockBlock } from './models/IUnlockBlock';
import { Transaction } from './transaction';



export interface IFaucetRequestContext {
  faucetRequest: IFaucetRequest;
  poWBuffer: Buffer;
}

export class WalletService {
  private readonly client: BasicClient;

  constructor(basicClient: BasicClient) {
    this.client = basicClient;
  }

  public async getFunds(address: string, color: string): Promise<bigint> {

    const unspents = await this.client.unspentOutputs({ addresses: [address] });
    const currentUnspent = unspents.unspentOutputs.find((x) => x.address.base58 == address);

    const balance = currentUnspent.outputs
      .filter(
        (o) =>
          ['ExtendedLockedOutputType', 'SigLockedColoredOutputType'].includes(o.output.type) &&
          typeof o.output.output.balances[color] != 'undefined'
      )
      .map((uid) => uid.output.output.balances)
      .reduce((balance: bigint, output) => (balance += BigInt(output[color])), BigInt(0));

    return balance;
  }

  public async getFaucetRequest(address: string): Promise<IFaucetRequestContext> {
    const manaPledge = await this.client.getAllowedManaPledge();

    const allowedManagePledge = manaPledge.accessMana.allowed[0];
    const consenseusManaPledge = manaPledge.consensusMana.allowed[0];

    const body: IFaucetRequest = {
      accessManaPledgeID: allowedManagePledge,
      consensusManaPledgeID: consenseusManaPledge,
      address: address,
      nonce: -1
    };

    const poWBuffer = Faucet.ToBuffer(body);

    const result: IFaucetRequestContext = {
      poWBuffer: poWBuffer,
      faucetRequest: body
    };

    return result;
  }

  public async sendOnLedgerRequest(keyPair: IKeyPair, address: string, chainId: string, payload: IOnLedger, transfer: bigint = 1n) {
    if (transfer <= 0) {
      transfer = 1n;
    }

    const wallet = new BasicWallet(this.client);

    const unspents = await wallet.getUnspentOutputs(address);
    const consumedOutputs = wallet.determineOutputsToConsume(unspents, transfer);
    console.log("consumedOutputs", consumedOutputs)
    const { inputs, consumedFunds } = wallet.buildInputs(consumedOutputs);
    console.log("address", address, "chainId", chainId, "transfer", transfer, "consumedFunds", consumedFunds);
    const outputs = wallet.buildOutputs(address, chainId, transfer, consumedFunds);

    const unlockBlocks: IUnlockBlock[] = [];

    const tx: ITransaction = {
      version: 0,
      timestamp: BigInt(Date.now()) * 1000000n,
      aManaPledge: Base58.encode(Buffer.alloc(32)),
      cManaPledge: Base58.encode(Buffer.alloc(32)),
      inputs: inputs,
      outputs: outputs,
      chainId: chainId,
      payload: OnLedger.ToBuffer(payload),
      unlockBlocks: []
    };

    const txEssence = Transaction.essence(tx, Buffer.alloc(0));

    const addressByOutputID: { [outputID: string]: string; } = {};
    for (const address in consumedOutputs) {
      for (const outputID in consumedOutputs[address]) {
        addressByOutputID[outputID] = address;
      }
    }

    const existingUnlockBlocks: { [address: string]: number; } = {};
    for (const index in inputs) {
      const addr = address == addressByOutputID[inputs[index]];
      if (addr) {
        if (existingUnlockBlocks[address] !== undefined) {
          unlockBlocks.push({ type: 1, referenceIndex: existingUnlockBlocks[address], publicKey: Buffer.alloc(0), signature: Buffer.alloc(0) });
          continue;
        }

        const signatureUnlockBlock = { type: 0, referenceIndex: 0, publicKey: keyPair.publicKey, signature: Transaction.sign(keyPair, txEssence) };
        existingUnlockBlocks[address] = unlockBlocks.length;
        unlockBlocks.push(signatureUnlockBlock);
      }
    }

    tx.unlockBlocks = unlockBlocks;

    const result = Transaction.bytes(tx, txEssence);

    const response = await this.client.sendTransaction({
      txn_bytes: result.toString("base64")
    });

    return response;
  }
}
