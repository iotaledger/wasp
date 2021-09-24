import { derived, get, Readable, Writable, writable } from 'svelte/store';
import type { IRound } from './models/IRound';
import type { Buffer, IKeyPair } from './wasp_client';
import { Base58 } from './wasp_client/crypto/base58';

export const isInitialized: Writable<boolean> = writable(false);

export const seed: Writable<Buffer> = writable()
export const seedString: Readable<string> = derived(seed, $seed => Base58.encode($seed))
export const keyPair: Writable<IKeyPair> = writable()
export const address: Writable<string> = writable()
export const addressIndex: Writable<number> = writable(0)
export const balance: Writable<bigint> = writable(0n)

export const timestamp: Writable<number> = writable()
export const isWorking: Writable<boolean> = writable()
export const requestingFunds: Writable<boolean> = writable(false)
export const placingBet: Writable<boolean> = writable(false)

export const START_GAME_STATE = "START_GAME";
export const GAME_RUNNING_STATE = "GAME_RUNNING";

export const state: Writable<string> = writable(START_GAME_STATE);
export const fundsRequested: Writable<boolean> = writable(false);
export const newAddressNeeded: Writable<boolean> = writable(false);
const RESET_ROUND: IRound = {
    active: false,
    logs: [],
    players: [],
    betSelection: undefined,
    betAmount: undefined,
    winningNumber: undefined,
    startedAt: undefined,
    number: undefined,
}

export const round: Writable<IRound> = writable(RESET_ROUND);

export function resetRound(): void {
    round.set(RESET_ROUND)
}

export function updateGameState(): void {
    if (get(balance) > 0n) {
        fundsRequested.set(true);
    }
    if (get(round).active) {
        state.set(GAME_RUNNING_STATE)
    }
    else {
        state.set(START_GAME_STATE)
    }
}

// balance.subscribe(_balance => {
//     console.log("balance", _balance)
// })