import type { Writable } from 'svelte/store';
import type { ILog } from "./ILog";
import type { IPlayer } from "./IPlayer";

export interface IRound {
    active: boolean
    logs: ILog[]
    players: IPlayer[]
    betSelection: number
    betAmount: bigint
    winningNumber: bigint
    startedAt: number
    number: bigint
}