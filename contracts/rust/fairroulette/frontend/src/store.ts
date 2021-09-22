import { derived, Readable, Writable, writable } from 'svelte/store';
import type { ILog } from "./models/ILog";
import type { IPlayer } from "./models/IPlayer";
import type { Buffer, IKeyPair } from './wasp_client';
import { Base58 } from "./wasp_client/crypto/base58";

export const seed: Writable<Buffer> = writable()
export const seedString: Readable<string> = derived(seed, $seed => Base58.encode($seed))
export const keyPair: Writable<IKeyPair> = writable()
export const address: Writable<string> = writable()
export const addressIndex: Writable<number> = writable(0)
export const balance: Writable<bigint> = writable(0n)
export const requestingFunds: Writable<boolean> = writable(false)
export const logs: Writable<ILog[]> = writable([])
export const players: Writable<IPlayer[]> = writable([])
export const selectedBetNumber: Writable<number> = writable()
export const selectedBetAmount: Writable<bigint> = writable()
export const startedAt: Writable<number> = writable()
export const winningNumber: Writable<bigint> = writable()