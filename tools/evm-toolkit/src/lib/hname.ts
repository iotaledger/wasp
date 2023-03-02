import { Bech32, Blake2b } from '@iota/crypto.js';
import { SimpleBufferCursor } from './simple_buffer_cursor';
import { Buffer } from 'buffer';
import { Converter } from "@iota/util.js";

export function hNameFromString(name): Number {
  const ScHNameLength = 4;
  const stringBytes = Converter.utf8ToBytes(name);
  const hash = Blake2b.sum256(stringBytes);

  for (let i = 0; i < hash.length; i += ScHNameLength) {
    const slice = hash.slice(i, i + ScHNameLength);
    const cursor = new SimpleBufferCursor(Buffer.from(slice));

    return cursor.readUInt32LE();
  }

  return 0;
}