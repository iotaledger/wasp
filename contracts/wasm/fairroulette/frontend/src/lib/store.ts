import { derived, get, Readable, Writable, writable } from 'svelte/store';
import { FairRouletteService } from './fairroulette_client';
import type { IRound } from './models/IRound';
import type { Buffer, IKeyPair } from './wasp_client';
import { Base58 } from './wasp_client/crypto/base58';

export enum BettingStep {
  NumberChoice = 1,
  AmountChoice = 2,
}

const RESET_ROUND: IRound = {
  active: false,
  logs: [],
  players: [],
  betSelection: undefined,
  betAmount: undefined,
  betPlaced: false,
  winningNumber: undefined,
  startedAt: undefined,
  number: undefined,
  winners: 0,
};

export const seed: Writable<Buffer> = writable();
export const seedString: Readable<string> = derived(seed, ($seed) => Base58.encode($seed));
export const keyPair: Writable<IKeyPair> = writable();
export const address: Writable<string> = writable();
export const addressIndex: Writable<number> = writable(0);
export const balance: Writable<bigint> = writable(0n);

export const round: Writable<IRound> = writable(RESET_ROUND);

export const timestamp: Writable<number> = writable();
export const timeToFinished: Readable<number | undefined> = derived(timestamp, ($timestamp) =>
  $timestamp && get(receivedRoundStarted) ? calculateRoundLengthLeft($timestamp) : undefined,
);

export const placingBet: Writable<boolean> = writable(false);
export const showBettingSystem: Writable<boolean> = writable(false);
export const bettingStep: Writable<BettingStep> = writable(1);

export const showWinningNumber: Writable<boolean> = writable(false);

export const firstTimeRequestingFunds: Writable<boolean> = writable(true);
export const requestingFunds: Writable<boolean> = writable(false);

export const isAWinnerPlayer: Writable<boolean> = writable(false);

export const addressesHistory: Writable<string[]> = writable([]);

// Added to bugfix system clocks unsynced,
// we can only rely on timeToFinished if the user received a roundStarted event
export const receivedRoundStarted: Writable<boolean> = writable(false);

export function resetRound(): void {
  receivedRoundStarted.set(false);
  round.set({ ...RESET_ROUND, winningNumber: get(round)?.winningNumber, players: [], logs: get(round)?.logs });
}

export function showWinnerAnimation(): void {
  isAWinnerPlayer.set(true);
  setTimeout(() => {
    isAWinnerPlayer.set(false);
  }, 20000);
}

export function resetBettingSystem(): void {
  showBettingSystem.set(false);
  bettingStep.set(BettingStep.NumberChoice);
  round.update(($round) => ({ ...$round, betSelection: undefined, betAmount: undefined }));
}

export function calculateRoundLengthLeft(timestamp: number): number | undefined {
  const roundStartedAt = get(round).startedAt;

  if (!timestamp || !roundStartedAt) return undefined;

  if (roundStartedAt == 0) {
    return 0;
  }

  const diff = Math.round(timestamp - roundStartedAt);

  const roundTimeLeft = Math.round(FairRouletteService.roundLength - diff);

  if (roundTimeLeft <= 0) {
    return 0;
  }
  return roundTimeLeft;
}
