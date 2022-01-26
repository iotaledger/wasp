import { Buffer } from "../../buffer";
import { SimpleBufferCursor } from "../utils/simple_buffer_cursor";

export interface OnLedgerArgument {
    key: string;
    value: number;
}

export interface IOnLedger {
    contract?: number;
    entrypoint?: number;
    arguments?: OnLedgerArgument[];
    nonce?: number;
}

export class OnLedgerHelper {
    public static ToBuffer(req: IOnLedger): Buffer {
        const buffer = new SimpleBufferCursor(Buffer.alloc(0));

        buffer.writeUInt32LE(0);
        buffer.writeUInt32LE(req.contract || 0);
        buffer.writeUInt32LE(req.entrypoint || 0);
        buffer.writeBytes(Buffer.alloc(1, 0));

        buffer.writeUInt32LE(req.arguments?.length || 0);

        if (req.arguments) {
            req.arguments.sort((lhs, rhs) => lhs.key.localeCompare(rhs.key));
            for (const arg of req.arguments) {
                const keyBuffer = Buffer.from(arg.key);

                buffer.writeUInt16LE(keyBuffer.length);
                buffer.writeBytes(keyBuffer);

                const valueBuffer = Buffer.alloc(8);
                valueBuffer.writeInt32LE(arg.value, 0);

                buffer.writeUInt32LE(valueBuffer.length);
                buffer.writeBytes(valueBuffer);
            }
        }

        return buffer.buffer;
    }
}
