import type { IResponse } from '../../api_common/response_models';

export interface IAllowedManaPledgeResponse extends IResponse {
    accessMana: {
        isFilterEnabled: boolean;
        allowed?: Array<string>;
    };

    consensusMana: {
        isFilterEnabled: boolean;
        allowed?: Array<string>;
    };
}
