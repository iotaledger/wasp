import config from '../../../config.dev';
import { BasicClient, Buffer, Colors, IKeyPair, IOffLedger, IOnLedger, OffLedger, WalletService } from '../wasp_client';
import { createNanoEvents, Emitter, Unsubscribe } from 'nanoevents';
import { HName } from '../wasp_client/crypto/hname';

type MessageHandlers = { [key: string]: (index: number) => void };
type ParameterResult = { [key: string]: Buffer };

export interface Bet {
  better: string;
  amount: number;
  betNumber: number;
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
  public static readonly roundTimeLeft: string = 'roundTimeLeft';
}

export class FairRouletteService {
  private readonly scName: string = config.contractName;
  private readonly scHName: string = HName.HashAsString(this.scName);
  private readonly scPlaceBet: string = 'placeBet';

  private client: BasicClient;
  private walletService: WalletService;
  private webSocket: WebSocket;
  private emitter: Emitter;

  public chainId: string;
  public static readonly roundLength: number = 60; // in seconds

  constructor(client: BasicClient, chainId: string) {
    this.walletService = new WalletService(client);
    this.client = client;
    this.chainId = chainId;
    this.emitter = createNanoEvents();

    this.connectWebSocket();
  }

  private connectWebSocket(): void {
    const webSocketUrl = config.waspWebSocketUrl.replace('%chainId', this.chainId);
    // eslint-disable-next-line no-console
    console.log(`Connecting to Websocket => ${webSocketUrl}`);
    this.webSocket = new WebSocket(webSocketUrl);
    this.webSocket.addEventListener('message', (x) => this.handleIncomingMessage(x));
    this.webSocket.addEventListener('close', () => setTimeout(this.connectWebSocket.bind(this), 1000));
  }

  private handleVmMessage(message: string[]): void {
    const messageHandlers: MessageHandlers = {
      'fairroulette.bet': () => {
        const bet: Bet = {
          better: message[2],
          amount: Number(message[3]),
          betNumber: Number(message[4]),
        };

        this.emitter.emit('betPlaced', bet);
      },

      'fairroulette.payout': (index) => {
        const bet: Bet = {
          better: message[2],
          amount: Number(message[3]),
          betNumber: undefined,
        };

        this.emitter.emit('payout', bet);
      },

      'fairroulette.round': (index) => {
        this.emitter.emit('roundNumber', message[2] || 0);
      },

      'fairroulette.start': (index) => {
          this.emitter.emit('roundStarted', message[1] || 0);
      },

      'fairroulette.stop': (index) => {
           this.emitter.emit('roundStopped');
      },

      'fairroulette.winner': (index) => {
        this.emitter.emit('winningNumber', message[2] || 0);
      },
    };

    const topic = message[0];
    if (typeof messageHandlers[topic] != 'undefined') {
      messageHandlers[topic](0);
    }
  }

  private handleIncomingMessage(message: MessageEvent<string>): void {
    // expect vmmsg <chain ID> <contract hname> contract.event|param1|param2|...
    const msg = message.data.toString().split(' ');
    if (msg.length != 4 || msg[0] != 'vmmsg') {
      return;
    }
    const topics = msg[3].split('|');
    this.handleVmMessage(topics);
  }

  public async placeBetOffLedger(keyPair: IKeyPair, betNumber: number, take: bigint): Promise<void> {
    let betRequest: IOffLedger = {
      requestType: 1,
      arguments: [{ key: '-number', value: betNumber }],
      balances: [{ balance: take, color: Colors.IOTA_COLOR_BYTES }],
      contract: HName.HashAsNumber(this.scName),
      entrypoint: HName.HashAsNumber(this.scPlaceBet),
      nonce: BigInt(performance.now() + performance.timeOrigin * 10000000),
    };

    betRequest = OffLedger.Sign(betRequest, keyPair);

    await this.client.sendOffLedgerRequest(this.chainId, betRequest);
    await this.client.sendExecutionRequest(this.chainId, OffLedger.GetRequestId(betRequest));
  }

  public async placeBetOnLedger(keyPair: IKeyPair, address: string, betNumber: number, take: bigint): Promise<void> {
    const betRequest: IOnLedger = {
      contract: HName.HashAsNumber(this.scName),
      entrypoint: HName.HashAsNumber(this.scPlaceBet),
      arguments: [
        {
          key: '-number',
          value: betNumber,
        },
      ],
    };

    await this.walletService.sendOnLedgerRequest(keyPair, address, this.chainId, betRequest, take);
  }

  public async callView(viewName: string): Promise<ParameterResult> {
    const response = await this.client.callView(this.chainId, this.scHName, viewName);
    const resultMap: ParameterResult = {};

    if (response.Items) {
      for (const item of response.Items) {
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

  public async getRoundTimeLeft(): Promise<number> {
    const response = await this.callView(ViewEntrypoints.roundTimeLeft);
    const roundTimeLeft = response[ViewEntrypoints.roundTimeLeft];

    if (!roundTimeLeft) {
      throw Error(`Failed to get ${ViewEntrypoints.roundTimeLeft}`);
    }

    return roundTimeLeft.readInt32LE(0);
  }

  public on<E extends keyof Events>(event: E, callback: Events[E]): Unsubscribe {
    return this.emitter.on(event, callback);
  }
}
