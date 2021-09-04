import { Writable, writable } from 'svelte/store'
import type { Buffer } from './client/buffer';
import type { IKeyPair } from './client/crypto/models/IKeyPair';

export const seed: Writable<Buffer> = writable();
export const keyPair: Writable<IKeyPair> = writable();
export const address: Writable<string> = writable();
export const addressIndex: Writable<number> = writable(0);
