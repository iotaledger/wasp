import { Buffer } from './buffer'
import { Faucet } from './binary_models/faucet_request'
import { OffLedger } from './binary_models/off_ledger'
import type { IAllowedManaPledgeResponse } from "./models/IAllowedManaResponse";
import type { IResponse } from "./models/IResponse";
import type { IFaucetRequest } from "./binary_models/IFaucetRequest";
import type { IFaucetResponse } from "./models/IFaucetResponse";
import type { IUnspentOutputsRequest } from "./models/IUnspentOutputsRequest";
import type { IUnspentOutputsResponse } from "./models/IUnspentOutputsResponse";
import type { ISendTransactionRequest } from "./models/ISendTransactionRequest";
import type { ISendTransactionResponse } from "./models/ISendTransactionResponse";
import type { IOffLedger } from "./binary_models/IOffLedger";
import type { IOffLedgerRequest } from "./models/IOffLedgerRequest";

export interface IExtendedResponse<U> {
  body: U;
  response: Response;
}

export interface BasicClientConfiguration {
  WaspAPIUrl: string;
  GoShimmerAPIUrl: string;
  SeedUnsafe: Buffer;
}

export interface IFaucetRequestContext {
  faucetRequest: IFaucetRequest,
  poWBuffer: any
}

export interface CallViewResponse extends IResponse {
  Items: [{ Key: string, Value: any }];
}


export class Colors {
  public static IOTA_COLOR_STRING = '11111111111111111111111111111111';
  public static IOTA_COLOR_BYTES = Buffer.alloc(32);
}

export class BasicClient {

  private configuration: BasicClientConfiguration;

  constructor(configuration: BasicClientConfiguration) {
    this.configuration = configuration;
  }

  public async allowedManaPledge(): Promise<IAllowedManaPledgeResponse> {
    return this.sendRequest<null, IAllowedManaPledgeResponse>(this.configuration.GoShimmerAPIUrl,
      "get", "mana/allowedManaPledge");
  }

  public async getFaucetRequest(address: string): Promise<IFaucetRequestContext> {
    const manaPledge = await this.allowedManaPledge();

    const allowedManagePledge = manaPledge.accessMana.allowed[0];
    const consenseusManaPledge = manaPledge.consensusMana.allowed[0];

    const body: IFaucetRequest = {
      accessManaPledgeID: allowedManagePledge,
      consensusManaPledgeID: consenseusManaPledge,
      address: address,
      nonce: -1
    };

    const poWBuffer = Faucet.ToBuffer(body);

    const result: IFaucetRequestContext = {
      poWBuffer: poWBuffer,
      faucetRequest: body
    };

    return result;
  }

  public async sendFaucetRequest(faucetRequest: IFaucetRequest) {
    const response = await this.sendRequest<IFaucetRequest, IFaucetResponse>(this.configuration.GoShimmerAPIUrl, "post", "faucet", faucetRequest);

    return response;
  }

  public async sendOffLedgerRequest(chainId: string, offLedgerRequest: IOffLedger) {
    const request = { Request: OffLedger.ToBuffer(offLedgerRequest).toString('base64') };
    const response = await this.sendRequestExt<IOffLedgerRequest, null>(this.configuration.WaspAPIUrl, "post", `request/${chainId}`, request);

    console.log(response.response);

    return response;
  }

  public async sendExecutionRequest(chainId: string, offLedgerRequestId: string) {
    const response = await this.sendRequestExt<IOffLedgerRequest, null>(this.configuration.WaspAPIUrl, "get", `chain/${chainId}/request/${offLedgerRequestId}/wait`);

    return response;
  }

  public async getFunds(address: string, color: string): Promise<bigint> {

    const unspents = await this.unspentOutputs({ addresses: [address] });
    const currentUnspent = unspents.unspentOutputs.find((x) => x.address.base58 === address);

    const balance = currentUnspent.outputs
      .filter(
        (o) =>
          ['ExtendedLockedOutputType', 'SigLockedColoredOutputType'].includes(o.output.type) &&
          typeof o.output.output.balances[color] !== 'undefined'
      )
      .map((uid) => uid.output.output.balances)
      .reduce((balance: bigint, output) => (balance += BigInt(output[color])), BigInt(0));

    return balance;
  }

  public async callView(chainId: string, contractHName: string, entryPoint: string): Promise<CallViewResponse> {
    const url = `chain/${chainId}/contract/${contractHName}/callview/${entryPoint}`;

    const result = await this.sendRequestExt<any, CallViewResponse>(this.configuration.WaspAPIUrl, 'get', url);

    return result.body;
  }

  public async unspentOutputs(request: IUnspentOutputsRequest): Promise<IUnspentOutputsResponse> {
    return this.sendRequest<IUnspentOutputsRequest, IUnspentOutputsResponse>(this.configuration.GoShimmerAPIUrl,
      "post", "ledgerstate/addresses/unspentOutputs", request);
  }

  public async sendTransaction(request: ISendTransactionRequest): Promise<ISendTransactionResponse> {
    return this.sendRequest<ISendTransactionRequest, ISendTransactionResponse>(this.configuration.GoShimmerAPIUrl,
      "post", "ledgerstate/transactions", request);
  }

  private async sendRequest<T, U extends IResponse>(
    url: string,
    verb: "put" | "post" | "get" | "delete",
    path: string,
    request?: T | undefined): Promise<U> {

    const response = await this.sendRequestExt<T, U>(url, verb, path, request);

    return response.body;
  }

  private async sendRequestExt<T, U extends IResponse>(
    url: string,
    verb: "put" | "post" | "get" | "delete",
    path: string,
    request?: T | undefined): Promise<IExtendedResponse<U>> {

    let response: U;
    let fetchResponse: Response;

    try {
      const headers: { [id: string]: string } = { };
      headers["Content-Type"] = "application/json";


      if (verb === "get") {
        fetchResponse = await fetch(
          `${url}/${path}`,
          {
            method: "get",
            headers,
          }
        );
      } else if (verb === "post") {
        fetchResponse = await fetch(
          `${url}/${path}`,
          {
            method: "post",
            headers,
            body: JSON.stringify(request)
          }
        );
      } else if (verb === "put") {
        fetchResponse = await fetch(
          `${url}/${path}`,
          {
            method: "put",
            headers,
            body: JSON.stringify(request)
          }
        );
      } else if (verb === "delete") {
        fetchResponse = await fetch(
          `${url}/${path}`,
          {
            method: "delete",
            headers
          }
        );
      }

      if (!fetchResponse) {
        throw new Error("No data was returned from the API");
      } else {
        try {
          response = await fetchResponse.json();
        } catch (err) {

          if (!fetchResponse.ok) {
            const text = await fetchResponse.text();
            throw new Error(err.message + "   ---   " + text);

          }
        }
        if (!fetchResponse.ok) {
          if (response.error) {
            throw new Error(response.error);
          } else {
            const isComError = fetchResponse.status >= 500;
            let msg = fetchResponse.statusText;

            if (msg === "Network Error") {
              msg = "There is a problem communicating with the network";
            }

            if (!msg.endsWith(".")) {
              msg += ".";
            }

            if (isComError) {
              msg += "\n\nPlease try again later.";
            }

            if (fetchResponse.status === 400) {
              msg = "";
              msg += response;
            }

            throw new Error(msg);
          }
        }
      }
    } catch (err) {
      throw new Error(`The application is not able to complete the request, due to the following error:\n\n${err.message}`);
    }

    return { body: response, response: fetchResponse };
  }
}
