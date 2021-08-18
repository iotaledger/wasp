import { Buffer } from './client/buffer';
import { BasicClient, Colors } from "./client/basic_client";
import type { IOffLedger } from "./client/binary_models/IOffLedger";
import { OffLedger } from "./client/binary_models/off_ledger";
import { HName } from "./client/crypto/hname";
import type { IKeyPair } from './client/crypto/models/IKeyPair';

export class FairRoulette {

  private readonly scName: string = 'fairroulete';
  private readonly scPlaceBet: string = 'placeBet';

  private client: BasicClient;

  public chainId: string;

  constructor(client: BasicClient, chainId: string) {
    this.client = client;
    this.chainId = chainId;
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

    console.log(betRequest)

    await this.client.sendOffLedgerRequest(this.chainId, betRequest);
    await this.client.sendExecutionRequest(this.chainId, OffLedger.GetId(betRequest));
  }
}
