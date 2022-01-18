import { Buffer } from '../buffer';
import { SimpleBufferCursor } from '../simple_buffer_cursor';
import type { IOnLedger } from './IOnLedger';

export class OnLedger {
  public static ToStruct(buffer: Buffer): IOnLedger {
    const reader = new SimpleBufferCursor(buffer);

    const contract = reader.readUInt32LE();
    const entrypoint = reader.readUInt32LE();
    const numArguments = reader.readUInt32LE();

    const args = [];
    for (let i = 0; i < numArguments; i++) {
      const sz16 = reader.readUInt16LE();
      const key = reader.readBytes(sz16);
      const sz32 = reader.readUInt32LE();
      const value = reader.readBytes(sz32);
      args.push({ key: key, value: value });
    }

    const offLedgerStruct: IOnLedger = {
      contract: contract,
      entrypoint: entrypoint,
      arguments: args,
      noonce: 0,
    };

    return offLedgerStruct;
  }

  public static ToBuffer(req: IOnLedger): Buffer {
    const buffer = new SimpleBufferCursor(Buffer.alloc(0));

    buffer.writeUInt32LE(0);
    buffer.writeUInt32LE(req.contract);
    buffer.writeUInt32LE(req.entrypoint);
    buffer.writeBytes(Buffer.alloc(1, 0));

    buffer.writeUInt32LE(req.arguments.length || 0);
    if (req.arguments) {
      req.arguments.sort((lhs,rhs)=> lhs.key.localeCompare(rhs.key));
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
