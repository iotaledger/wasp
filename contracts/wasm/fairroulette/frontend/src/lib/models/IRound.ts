import type { ILog } from './ILog';
import type { IPlayer } from './IPlayer';

export interface IRound {
  active: boolean;
  logs: ILog[];
  players: IPlayer[];
  betSelection: number | undefined;
  betAmount: bigint | undefined;
  winningNumber: bigint | undefined;
  startedAt: number | undefined;
  number: bigint | undefined;
  betPlaced: boolean;
  winners: number;
}
