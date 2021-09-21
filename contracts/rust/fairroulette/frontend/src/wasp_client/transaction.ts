import { Base58 } from './crypto/base58';
import { Buffer } from './buffer';
import { Colors } from './colors';
import { ED25519 } from './crypto/ed25519';
import type { ITransaction } from './models/ITransaction';
import type { IKeyPair } from "./models";

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
        const buffers: Buffer[] = [];

        // version
        const version = Buffer.alloc(1);
        version.writeUInt8(tx.version, 0);
        buffers.push(version);

        // timestamp
        const timestamp = Buffer.alloc(8);
        timestamp.writeBigInt64LE(tx.timestamp, undefined);
        buffers.push(timestamp);

        // aManaPledge
        buffers.push(Base58.decode(tx.aManaPledge));

        // cManaPledge
        buffers.push(Base58.decode(tx.cManaPledge));

        // Input Size
        const inputsCount = Buffer.alloc(2);
        inputsCount.writeUInt16LE(tx.inputs.length, undefined);
        buffers.push(inputsCount);

        // Inputs
        for (const input of tx.inputs) {
            const inputType = Buffer.alloc(1);
            inputType.writeUInt8(0, undefined);
            buffers.push(inputType);

            const decodedInput = Base58.decode(input);
            buffers.push(decodedInput);
        }

        // Output count
        const outputsCount = Buffer.alloc(2);
        outputsCount.writeUInt16LE(Object.keys(tx.outputs).length, undefined);
        buffers.push(outputsCount);



        // Outputs
        const bufferOutputs: Buffer[] = [];
        for (const address in tx.outputs) {
            const outputType = Buffer.alloc(1);
            outputType.writeUInt8(3, undefined);

            const balancesCount = Buffer.alloc(4);
            balancesCount.writeUInt32LE(tx.outputs[address].length, undefined);

            const bufferColors: Buffer[] = [];
            for (const balance of tx.outputs[address]) {
                const colorValueBuffer = Buffer.alloc(8);
                colorValueBuffer.writeBigUInt64LE(balance.value, undefined);
                bufferColors.push(Buffer.concat([Colors.IOTA_COLOR_BYTES, colorValueBuffer]));
            }
            bufferColors.sort((a, b) => a.compare(b));

            const decodedAddress = Base58.decode(address);

            const type = Buffer.alloc(1);

            if (address == tx.chainId) {
                type.writeInt8(4, undefined); // no timelock, no fallbackAddress, HAS payload
            }

            let output = Buffer.concat([outputType, balancesCount, Buffer.concat(bufferColors), decodedAddress, type]);

            if (address == tx.chainId) {
                const payloadLength = Buffer.alloc(2);
                payloadLength.writeUInt16LE(tx.payload.length, undefined);

                output = Buffer.concat([output, payloadLength, tx.payload]);
            }

            bufferOutputs.push(output);
        }

        bufferOutputs.sort((a, b) => a.compare(b));
        buffers.push(Buffer.concat(bufferOutputs));
        buffers.push(Buffer.alloc(4));

        return Buffer.concat(buffers);
    }

    /**
     * Get the bytes for a transaction.
     * @param tx The tx to get the bytes for.
     * @param essence Existing essence.
     * @returns The bytes of the transaction.
     */
    public static bytes(tx: ITransaction, essence?: Buffer): Buffer {
        const buffers: Buffer[] = [];

        const payloadType = Buffer.alloc(4);
        payloadType.writeUInt32LE(1337, undefined);
        buffers.push(payloadType);

        const essenceBytes = Transaction.essence(tx);
        buffers.push(essenceBytes);

        const unlockBlocksCount = Buffer.alloc(2);
        unlockBlocksCount.writeUInt16LE(tx.unlockBlocks.length, undefined);
        buffers.push(unlockBlocksCount);

        for (const index in tx.unlockBlocks) {
            const ubType = tx.unlockBlocks[index].type;

            const unlockBlockType = Buffer.alloc(1);
            unlockBlockType.writeUInt8(ubType, undefined);
            buffers.push(unlockBlockType);

            if (ubType === 0) {
                const ED25519Type = Buffer.alloc(1);
                ED25519Type.writeUInt8(0, undefined);
                buffers.push(ED25519Type);
                buffers.push(tx.unlockBlocks[index].publicKey);
                buffers.push(tx.unlockBlocks[index].signature);
                continue;
            }

            const referencedIndex = Buffer.alloc(2);
            referencedIndex.writeUInt16LE(tx.unlockBlocks[index].referenceIndex, undefined);
            buffers.push(referencedIndex);
        }

        const payloadSize = Buffer.concat(buffers).length;
        const payloadSizeBuffer = Buffer.alloc(4);
        payloadSizeBuffer.writeUInt32LE(payloadSize, undefined);
        buffers.unshift(payloadSizeBuffer);

        return Buffer.concat(buffers);
    }
}
