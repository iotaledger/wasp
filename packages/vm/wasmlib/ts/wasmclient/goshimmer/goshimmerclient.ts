import { Buffer } from '../buffer';
import { Configuration } from '../configuration';

import { IFaucetRequest, IFaucetRequestContext, IFaucetResponse } from './faucet/faucet_models';
import { FaucetHelper } from './faucet/faucet_helper';

import { IUnspentOutputsRequest, IUnspentOutputsResponse } from './models/unspent_outputs';
import { IAllowedManaPledgeResponse } from './models/mana';

import { PoWWorkerManager } from './pow_web_worker/pow_worker_manager';

import * as requestSender from '../api_common/request_sender';

interface GoShimmerClientConfiguration {
    APIUrl: string;
    SeedUnsafe: Buffer | null;
}

const IOTA_COLOR: string = 'IOTA';

export class GoShimmerClient {
    private readonly goShimmerConfiguration: GoShimmerClientConfiguration;
    private readonly powManager: PoWWorkerManager = new PoWWorkerManager();

    constructor(configuration: Configuration) {
        this.goShimmerConfiguration = { APIUrl: configuration.goShimmerApiUrl, SeedUnsafe: configuration.seed };
    }

    public async getIOTABalance(address: string): Promise<bigint> {
        const iotaBalance = await this.getBalance(address, IOTA_COLOR);
        return iotaBalance;
    }

    private async getBalance(address: string, color: string): Promise<bigint> {
        if (color == IOTA_COLOR) {
            color = '11111111111111111111111111111111';
        }

        const unspents = await this.unspentOutputs({ addresses: [address] });
        const currentUnspent = unspents.unspentOutputs.find((x) => x.address.base58 == address);

        const balance = currentUnspent!.outputs
            .filter(
                (o) =>
                    ['ExtendedLockedOutputType', 'SigLockedColoredOutputType'].includes(o.output.type) &&
                    typeof o.output.output.balances[color] != 'undefined'
            )
            .map((uid) => uid.output.output.balances)
            .reduce((balance: bigint, output) => (balance += BigInt(output[color])), BigInt(0));

        return balance;
    }

    private async unspentOutputs(request: IUnspentOutputsRequest): Promise<IUnspentOutputsResponse> {
        return requestSender.sendRequest<IUnspentOutputsRequest, IUnspentOutputsResponse>(
            this.goShimmerConfiguration.APIUrl,
            'post',
            'ledgerstate/addresses/unspentOutputs',
            request
        );
    }

    public async requestFunds(address: string): Promise<boolean> {
        try {
            const faucetRequestContext = await this.getFaucetRequest(address);
            const response = await this.sendFaucetRequest(faucetRequestContext.faucetRequest);
            const success = response.error === undefined && response.id !== undefined;
            return success;
        } catch (ex: unknown) {
            const error = ex as Error;
            console.error(error.message);
            return false;
        }
    }

    private async getFaucetRequest(address: string): Promise<IFaucetRequestContext> {
        const manaPledge = await this.getAllowedManaPledge();

        const allowedManagePledge = manaPledge.accessMana?.allowed ? manaPledge.accessMana.allowed[0] : '';
        const consenseusManaPledge = manaPledge.consensusMana?.allowed ? manaPledge.consensusMana?.allowed[0] : '';

        const body: IFaucetRequest = {
            accessManaPledgeID: allowedManagePledge,
            consensusManaPledgeID: consenseusManaPledge,
            address: address,
            nonce: -1,
        };

        const poWBuffer = FaucetHelper.ToBuffer(body);

        body.nonce = await this.powManager.requestProofOfWork(12, poWBuffer);

        const result: IFaucetRequestContext = {
            poWBuffer: poWBuffer,
            faucetRequest: body,
        };

        return result;
    }

    private async getAllowedManaPledge(): Promise<IAllowedManaPledgeResponse> {
        return requestSender.sendRequest<null, IAllowedManaPledgeResponse>(
            this.goShimmerConfiguration.APIUrl,
            'get',
            'mana/allowedManaPledge'
        );
    }

    private async sendFaucetRequest(faucetRequest: IFaucetRequest): Promise<IFaucetResponse> {
        const response = await requestSender.sendRequest<IFaucetRequest, IFaucetResponse>(
            this.goShimmerConfiguration.APIUrl,
            'post',
            'faucet',
            faucetRequest
        );
        return response;
    }
}
