import { Buffer } from "../../buffer";
import { SimpleBufferCursor } from "../utils/simple_buffer_cursor";
import { Base58, IKeyPair, ED25519 } from "../../crypto";

import type { BuiltOutputResult } from "../wallet/wallet";

import { IResponse } from "../../api_common/response_models";
import { Colors } from "../../colors";

export interface ISendTransactionRequest {
    txn_bytes: string;
}

export interface ISendTransactionResponse extends IResponse {
    transaction_id?: string;
}

export interface IUnlockBlock {
    type: number;
    referenceIndex: number;
    publicKey: Buffer;
    signature: Buffer;
}

export interface ITransaction {
    /**
     * The transaction's version.
     */
    version: number;

    /**
     * The transaction's timestamp.
     */
    timestamp: bigint;

    /**
     * The nodeID to pledge access mana.
     */
    aManaPledge: string;

    /**
     * The nodeID to pledge consensus mana.
     */
    cManaPledge: string;

    /**
     * The inputs to send.
     */
    inputs: string[];

    payload: Buffer;

    chainId: string;
    /**
     * The outputs to send.
     */
    outputs: BuiltOutputResult;

    /**
     * The signatures to send.
     */
    unlockBlocks: IUnlockBlock[];
}

/**
 * Class to help with transactions.
 */
export class Transaction {
    /**
     * Sign a transaction.
     * @param keyPair The key pair to sign with.
     * @param buffer The data to sign.
     * @returns The signature.
     */
    public static sign(keyPair: IKeyPair, buffer: Buffer): Buffer {
        return ED25519.privateSign(keyPair, buffer);
    }

    /**
     * Get the essence for a transaction.
     * @param tx The tx to get the essence for.
     * @returns The essence of the transaction.
     */
    public static essence(tx: ITransaction, payload: Buffer = Buffer.alloc(0)): Buffer {
        const essenceBuffer = new SimpleBufferCursor();

        essenceBuffer.writeInt8(tx.version);
        essenceBuffer.writeUInt64LE(tx.timestamp);
        essenceBuffer.writeBytes(Base58.decode(tx.aManaPledge));
        essenceBuffer.writeBytes(Base58.decode(tx.cManaPledge));

        essenceBuffer.writeUInt16LE(tx.inputs.length);
        for (const input of tx.inputs) {
            essenceBuffer.writeInt8(0);
            const decodedInput = Base58.decode(input);
            essenceBuffer.writeBytes(decodedInput);
        }

        essenceBuffer.writeUInt16LE(Object.keys(tx.outputs).length);

        const outputBuffers: SimpleBufferCursor[] = [];

        for (const address in tx.outputs) {
            const outputBuffer = new SimpleBufferCursor();

            outputBuffer.writeInt8(3);
            outputBuffer.writeUInt32LE(tx.outputs[address].length);

            const bufferColors: Buffer[] = [];

            for (const balance of tx.outputs[address]) {
                const colorValueBuffer = Buffer.alloc(8);
                colorValueBuffer.writeBigUInt64LE(balance.value, 0);
                bufferColors.push(Buffer.concat([Colors.IOTA_COLOR_BYTES, colorValueBuffer]));
            }

            bufferColors.sort((a, b) => a.compare(b));
            bufferColors.forEach((x) => outputBuffer.writeBytes(x));

            const decodedAddress = Base58.decode(address);
            outputBuffer.writeBytes(decodedAddress);

            if (address == tx.chainId) {
                outputBuffer.writeInt8(4);
                outputBuffer.writeUInt16LE(tx.payload.length);
                outputBuffer.writeBytes(tx.payload);
            } else {
                outputBuffer.writeInt8(0);
            }

            outputBuffers.push(outputBuffer);
        }

        outputBuffers.sort((a, b) => a.buffer.compare(b.buffer));
        outputBuffers.forEach((x) => essenceBuffer.writeBytes(x.buffer));

        essenceBuffer.writeUInt32LE(0);

        return essenceBuffer.buffer;
    }

    /**
     * Get the bytes for a transaction.
     * @param tx The tx to get the bytes for.
     * @param essence Existing essence.
     * @returns The bytes of the transaction.
     */
    public static bytes(tx: ITransaction): Buffer {
        const buffer = new SimpleBufferCursor();

        buffer.writeUInt32LE(1337);

        const essenceBytes = Transaction.essence(tx);
        buffer.writeBytes(essenceBytes);
        buffer.writeUInt16LE(tx.unlockBlocks.length);

        // TODO: Update this in the next refactoring.
        // eslint-disable-next-line @typescript-eslint/no-for-in-array
        for (const index in tx.unlockBlocks) {
            const ubType = tx.unlockBlocks[index].type;

            buffer.writeInt8(ubType);

            if (ubType === 0) {
                buffer.writeInt8(0);
                buffer.writeBytes(tx.unlockBlocks[index].publicKey);
                buffer.writeBytes(tx.unlockBlocks[index].signature);

                continue;
            }

            buffer.writeUInt16LE(tx.unlockBlocks[index].referenceIndex);
        }

        const returnBuffer = new SimpleBufferCursor();

        returnBuffer.writeUInt32LE(buffer.buffer.length);
        returnBuffer.writeBytes(buffer.buffer);

        return returnBuffer.buffer;
    }
}
