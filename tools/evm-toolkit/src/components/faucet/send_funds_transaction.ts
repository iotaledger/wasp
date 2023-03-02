import { Blake2b, Ed25519 } from "@iota/crypto.js";
import {
  ADDRESS_UNLOCK_CONDITION_TYPE,
  BASIC_OUTPUT_TYPE,
  Bech32Helper,
  DEFAULT_PROTOCOL_VERSION,
  ED25519_ADDRESS_TYPE,
  ED25519_SIGNATURE_TYPE,
  METADATA_FEATURE_TYPE,
  SENDER_FEATURE_TYPE,
  serializeTransactionEssence,
  SIGNATURE_UNLOCK_TYPE,
  TransactionHelper,
  TRANSACTION_ESSENCE_TYPE,
  TRANSACTION_PAYLOAD_TYPE,
  type IBasicOutput,
  type IBlock,
  type ITransactionEssence,
  type ITransactionPayload,
  type IUTXOInput,
  type UnlockTypes
} from "@iota/iota.js";
import { Converter, WriteStream } from "@iota/util.js";
import type { IotaWallet } from './iota_wallet';
import { SimpleBufferCursor } from '../../lib/simple_buffer_cursor';

export class SendFundsTransaction {
  private wallet: IotaWallet;

  constructor(client: IotaWallet) {
    this.wallet = client;
  }

  private createSendFundsMetadata(evmAddress: string, amount: bigint, gas: bigint) {
    const metadata = new SimpleBufferCursor();

    /* Write contract meta data */
    metadata.writeUInt32LE(0x0); // nil sender contract
    metadata.writeUInt32LE(0x3C4B5E02); // "accounts"
    metadata.writeUInt32LE(0x23F4E3A1); // "transferAllowanceTo"
    metadata.writeUInt64LE(gas); // gas

    /* Create evm address buffer */
    const evmAddressBuffer = new SimpleBufferCursor();
    evmAddressBuffer.writeInt8(3); // EVM address type (3)   
    evmAddressBuffer.writeUint8Array(Converter.hexToBytes(evmAddress.toLowerCase())); // EVM address


    /* Write contract arguments */
    metadata.writeUInt32LE(2);

    // Write evm address (arg1)
    metadata.writeUInt16LE(1);
    metadata.writeInt8("a".charCodeAt(0))
    metadata.writeUInt32LE(evmAddressBuffer.buffer.length);
    metadata.writeBytes(evmAddressBuffer.buffer);

    // Write account creation flag (arg2)
    metadata.writeUInt16LE(1);
    metadata.writeInt8("c".charCodeAt(0))
    metadata.writeUInt32LE(1);
    metadata.writeUInt8(1);

    /* Write allowance */
    metadata.writeUInt8(0); // Has allowance (255 if no allowance is set)
    metadata.writeUInt64LE(amount - gas); // IOTA amount to send
    metadata.writeUInt16LE(2); // Length of native assets data (we send no native assets, in this case we need to write two 0 bytes, and therefore provide the length of 2)
    metadata.writeUInt8(0); // Part of the empty native assets
    metadata.writeUInt8(0); // Part of the empty native assets
    metadata.writeUInt16LE(0); // Amount of NFTs

    return metadata.buffer;
  }

  public async sendFundsToEVMAddress(evmAddress: string, chainId: string, amount: bigint, gas: bigint): Promise<void> {
    const addressHex = Converter.bytesToHex(this.wallet.address.toAddress(), true);
    const addressBech32 = Bech32Helper.toBech32(ED25519_ADDRESS_TYPE, this.wallet.address.toAddress(), this.wallet.nodeInfo.protocol.bech32Hrp);

    const chainAddress = Bech32Helper.addressFromBech32(chainId, this.wallet.nodeInfo.protocol.bech32Hrp);

    const outputs = await this.wallet.indexer.basicOutputs({
      addressBech32: addressBech32
    });

    if (outputs.items.length == 0) {
      throw new Error("Could not find outputs to consume");
    }

    const outputId = outputs.items[0];
    const output = await this.wallet.client.output(outputId);

    if (output == null) {
      throw new Error("Could not fetch output data");
    }

    const metadata = this.createSendFundsMetadata(evmAddress, amount, gas);
    const metadataHex = Converter.bytesToHex(metadata, true);

    const basicOutput: IBasicOutput = {
      type: BASIC_OUTPUT_TYPE,
      amount: amount.toString(),
      nativeTokens: [
        // Add a list native tokens here
      ],
      unlockConditions: [
        {
          type: ADDRESS_UNLOCK_CONDITION_TYPE,
          address: chainAddress
        }
      ],
      features: [
        {
          type: SENDER_FEATURE_TYPE,
          address: {
            type: ED25519_ADDRESS_TYPE,
            pubKeyHash: addressHex,
          }
        },
        {
          type: METADATA_FEATURE_TYPE,
          data: metadataHex,
        }
      ]
    };

    const storageDeposit = TransactionHelper.getStorageDeposit(basicOutput, this.wallet.nodeInfo.protocol.rentStructure);
    amount = (amount - BigInt(storageDeposit));
    basicOutput.amount = amount.toString();


    const remainderBasicOutput: IBasicOutput = {
      type: BASIC_OUTPUT_TYPE,
      amount: (BigInt(output.output.amount) - amount).toString(),
      nativeTokens: [],
      unlockConditions: [
        {
          type: ADDRESS_UNLOCK_CONDITION_TYPE,
          address: {
            type: ED25519_ADDRESS_TYPE,
            pubKeyHash: addressHex
          }
        }
      ],
      features: []
    };

    const input: IUTXOInput = TransactionHelper.inputFromOutputId(outputId);
    const inputsCommitment = TransactionHelper.getInputsCommitment([output.output]);
    const protocolInfo = await this.wallet.client.protocolInfo();

    const transactionEssence: ITransactionEssence = {
      type: TRANSACTION_ESSENCE_TYPE,
      networkId: TransactionHelper.networkIdFromNetworkName(protocolInfo.networkName),
      inputs: [input],
      inputsCommitment,
      outputs: [basicOutput, remainderBasicOutput],
      payload: undefined
    };

    const wsTsxEssence = new WriteStream();
    serializeTransactionEssence(wsTsxEssence, transactionEssence);
    const essenceFinal = wsTsxEssence.finalBytes();
    const essenceHash = Blake2b.sum256(essenceFinal);

    const unlockCondition: UnlockTypes = {
      type: SIGNATURE_UNLOCK_TYPE,
      signature: {
        type: ED25519_SIGNATURE_TYPE,
        publicKey: Converter.bytesToHex(this.wallet.publicKey, true),
        signature: Converter.bytesToHex(this.wallet.sign(essenceHash), true)
      }
    };

    const transactionPayload: ITransactionPayload = {
      type: TRANSACTION_PAYLOAD_TYPE,
      essence: transactionEssence,
      unlocks: [unlockCondition]
    };

    const block: IBlock = {
      protocolVersion: DEFAULT_PROTOCOL_VERSION,
      parents: [],
      payload: transactionPayload,
      nonce: "0"
    };

    await this.wallet.client.blockSubmit(block);
  }
}