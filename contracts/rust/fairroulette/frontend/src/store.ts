import { derived, Readable, Writable, writable } from 'svelte/store';
import type { Buffer, IKeyPair } from './wasp_client';
import { Base58 } from "./wasp_client/crypto/base58";
import type {ILogData} from "./models/ILogEntries";
import type {IPlayerData} from "./models/IPlayerEntries";

export const seed: Writable<Buffer> = writable()
export const seedString: Readable<string> = derived(seed, $seed => Base58.encode($seed))
export const keyPair: Writable<IKeyPair> = writable()
export const address: Writable<string> = writable()
export const addressIndex: Writable<number> = writable(0)
export const balance: Writable<bigint> = writable(0n)
export const requestingFunds: Writable<boolean> = writable(false)
export const logs: Writable<ILogData[]> = writable([])
export const players: Writable<IPlayerData[]> = writable([])
export const selectedBetNumber: Writable<number> = writable(undefined)
export const selectedBetAmount: Writable<number> = writable(undefined)
export const startedAt: Writable<number> = writable(undefined)
export const winningNumber: Writable<bigint> = writable(undefined)