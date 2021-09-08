import config from '../../config.dev'
import { BasicClient, Colors } from '../client/basic_client'
import { Buffer } from '../client/buffer'
import { createNanoEvents, Emitter } from 'nanoevents'
import { HName } from '../client/crypto/hname'
import { OffLedger } from '../client/binary_models/off_ledger'
import type { IOffLedger } from '../client/binary_models/IOffLedger';
import type { IKeyPair } from '../client/crypto/models/IKeyPair';

type MessageHandlers = { [key: string]: (index: number) => void; };
type ParameterResult = { [key: string]: Buffer; };

export interface Bet {
  better: string,
  amount: number,
  betNumber: number,
}

export interface Events {
  roundStarted: (timestamp: number) => void;
  roundStopped: () => void;
  betPlaced: (bet: Bet) => void;
  roundNumber: (roundNr: bigint) => void;
  payout: (bet: Bet) => void;
  winningNumber: (number: bigint) => void;
}

export class ViewEntrypoints {
  public static readonly roundStartedAt: string = 'roundStartedAt';
  public static readonly roundNumber: string = 'roundNumber';
  public static readonly roundStatus: string = 'roundStatus';
  public static readonly lastWinningNumber: string = 'lastWinningNumber';
}

export class FairRoulette {
  private readonly scName: string = 'fairroulette';
  private readonly scHName: string = HName.HashAsString(this.scName);
  private readonly scPlaceBet: string = 'placeBet';

  private client: BasicClient;
  private webSocket: WebSocket;
  private emitter: Emitter;

  public chainId: string;
  public readonly roundLength: number = 30; // in seconds

  constructor(client: BasicClient, chainId: string) {
    this.client = client;
    this.chainId = chainId;
    this.emitter = createNanoEvents();

    const webSocketUrl = config.waspWebSocketUrl.replace("%chainId", chainId);

    this.webSocket = new WebSocket(webSocketUrl);
    this.webSocket.addEventListener("message", x => this.handleIncomingMessage(x));
  }

  private handleVmMessage(message: string[]): void {
    const messageHandlers: MessageHandlers = {
      'fairroulette.bet.placed': (index) => {
        const bet: Bet = {
          better: message[index + 1],
          amount: Number(message[index + 2]),
          betNumber: Number(message[index + 3])
        };

        this.emitter.emit('betPlaced', bet);
      },

      'fairroulette.round.state': (index) => {
        if (message[index + 1] == '1') {
          this.emitter.emit('roundStarted', message[index + 2]);
        } else {
          this.emitter.emit('roundStopped');
        }
      },

      'fairroulette.round.number': (index) => {
        this.emitter.emit('roundNumber', message[index + 1] || 0);
      },

      'fairroulette.round.winning_number': (index) => {
        this.emitter.emit('winningNumber', message[index + 1] || 0);
      },

      'fairroulette.payout': (index) => {
        const bet: Bet = {
          better: message[index + 1],
          amount: Number(message[index + 2]),
          betNumber: -1,
        };

        this.emitter.emit('payout', bet);
      }
    };

    const topicIndex = 3;
    const topic = message[topicIndex];

    if (typeof messageHandlers[topic] != 'undefined') {
      messageHandlers[topic](topicIndex);
    }
  }

  private handleIncomingMessage(message: MessageEvent<string>): void {
    const msg = message.data.split(' ');
    console.log(msg);

    if (msg.length == 0) {
      return;
    }

    if (msg[0] != 'vmmsg') {
      return;
    }


    this.handleVmMessage(msg);
  }

  public async placeBet(keyPair: IKeyPair, betNumber: number, take: number): Promise<void> {
    const tokenamount = Buffer.alloc(8);
    tokenamount.writeInt32LE(betNumber, 0);

    let betRequest: IOffLedger = {
      requestType: 1,
      arguments: [{ key: '-number', value: betNumber }],
      balances: [{ balance: BigInt(take), color: Colors.IOTA_COLOR_BYTES }],
      contract: HName.HashAsNumber(this.scName),
      entrypoint: HName.HashAsNumber(this.scPlaceBet),
      noonce: BigInt(performance.now() + performance.timeOrigin * 10000000),
    };

    betRequest = OffLedger.Sign(betRequest, keyPair);

    await this.client.sendOffLedgerRequest(this.chainId, betRequest);
    await this.client.sendExecutionRequest(this.chainId, OffLedger.GetRequestId(betRequest));
  }

  public async placeBetOnLedger(keyPair: IKeyPair, betNumber: number, take: number): Promise<void> {
    const tokenamount = Buffer.alloc(8);
    tokenamount.writeInt32LE(betNumber, 0);

    let betRequest: IOffLedger = {
      requestType: 1,
      arguments: [{ key: '-number', value: betNumber }],
      balances: [{ balance: BigInt(take), color: Colors.IOTA_COLOR_BYTES }],
      contract: HName.HashAsNumber(this.scName),
      entrypoint: HName.HashAsNumber(this.scPlaceBet),
      noonce: BigInt(performance.now() + performance.timeOrigin * 10000000),
    };

    betRequest = OffLedger.Sign(betRequest, keyPair);

    await this.client.sendOffLedgerRequest(this.chainId, betRequest);
    await this.client.sendExecutionRequest(this.chainId, OffLedger.GetRequestId(betRequest));
  }

  public async callView(viewName: string, args?: any): Promise<ParameterResult> {
    const response = await this.client.callView(this.chainId, this.scHName, viewName);
    const resultMap: ParameterResult = {};

    if (response.Items) {
      for (let item of response.Items) {
        const key = Buffer.from(item.Key, 'base64').toString();
        const value = Buffer.from(item.Value, 'base64');

        resultMap[key] = value;
      }
    }

    return resultMap;
  }

  public async getRoundStatus(): Promise<number> {
    const response = await this.callView(ViewEntrypoints.roundStatus);
    const roundStatus = response[ViewEntrypoints.roundStatus];

    if (!roundStatus) {
      throw Error(`Failed to get ${ViewEntrypoints.roundStatus}`);
    }

    return roundStatus.readUInt16LE(0);
  }

  public async getRoundNumber(): Promise<bigint> {
    const response = await this.callView(ViewEntrypoints.roundNumber);
    const roundNumber = response[ViewEntrypoints.roundNumber];

    if (!roundNumber) {
      throw Error(`Failed to get ${ViewEntrypoints.roundNumber}`);
    }

    return roundNumber.readBigUInt64LE(0);
  }

  public async getRoundStartedAt(): Promise<number> {
    const response = await this.callView(ViewEntrypoints.roundStartedAt);
    const roundStartedAt = response[ViewEntrypoints.roundStartedAt];

    if (!roundStartedAt) {
      throw Error(`Failed to get ${ViewEntrypoints.roundStartedAt}`);
    }

    return roundStartedAt.readInt32LE(0);
  }

  public async getLastWinningNumber(): Promise<bigint> {
    const response = await this.callView(ViewEntrypoints.lastWinningNumber);
    const lastWinningNumber = response[ViewEntrypoints.lastWinningNumber];

    if (!lastWinningNumber) {
      throw Error(`Failed to get ${ViewEntrypoints.lastWinningNumber}`);
    }

    return lastWinningNumber.readBigUInt64LE(0);
  }

  public on<E extends keyof Events>(event: E, callback: Events[E]) {
    return this.emitter.on(event, callback);
  }
}

