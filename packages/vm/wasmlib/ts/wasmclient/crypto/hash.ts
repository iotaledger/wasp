import {blake2b} from 'blakejs';
import {Buffer} from '../buffer';

export class Hash {
    public static from(bytes: Buffer): Buffer {
        // @ts-ignore
        return blake2b(bytes, undefined, 32 /* Blake256 */);
    }
}