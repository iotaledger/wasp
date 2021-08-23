import { Buffer } from './client/buffer';
import { BasicClient, Colors } from "./client/basic_client";
import type { IOffLedger } from "./client/binary_models/IOffLedger";
import { OffLedger } from "./client/binary_models/off_ledger";
import { HName } from "./client/crypto/hname";
import type { IKeyPair } from './client/crypto/models/IKeyPair';
import config from '../config.dev';
import { createNanoEvents, Emitter } from "nanoevents"

interface Events {
  start: (startedAt: number) => void
}

export class FairRoulette {

  private readonly scName: string = 'fairroulete';
  private readonly scPlaceBet: string = 'placeBet';

  private client: BasicClient;
  private webSocket: WebSocket;
  private emitter: Emitter;

  public chainId: string;

  constructor(client: BasicClient, chainId: string) {
    this.client = client;
    this.chainId = chainId;
    this.emitter = createNanoEvents();

    const webSocketUrl = config.waspWebSocketUrl.replace("%chainId", chainId);

    this.webSocket = new WebSocket(webSocketUrl);
    this.webSocket.addEventListener("message", this.handleIncomingMessage);
    this.webSocket.addEventListener("open", (ev) => {
      console.log("Opened")
    });
  }

  private handleIncomingMessage(message: MessageEvent<any>) {
    console.log(message.data);
  }

  public async placeBet(keyPair: IKeyPair, betNumber: number, take: number) {
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
    await this.client.sendExecutionRequest(this.chainId, OffLedger.GetId(betRequest));
  }

  public on<E extends keyof Events>(event: E, callback: Events[E]) {
    return this.emitter.on(event, callback)
  }
}
