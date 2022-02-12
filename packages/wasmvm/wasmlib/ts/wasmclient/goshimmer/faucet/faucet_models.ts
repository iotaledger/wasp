import type {IResponse} from '../../api_common/response_models';
import {Buffer} from '../../buffer';

export interface IFaucetRequest {
    accessManaPledgeID: string;
    consensusManaPledgeID: string;
    address: string;
    nonce: number;
}

export interface IFaucetResponse extends IResponse {
    id?: string;
}

export interface IFaucetRequestContext {
    faucetRequest: IFaucetRequest;
    poWBuffer: Buffer;
}
