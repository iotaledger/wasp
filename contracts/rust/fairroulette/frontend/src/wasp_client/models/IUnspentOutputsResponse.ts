import type { IResponse } from "./IResponse";

export interface IUnspentOutputsResponse extends IResponse {
    unspentOutputs: {
        address: {
            type: string;
            base58: string;
        };

        outputs: {
            output: {
                outputID: {
                    base58: string;
                    transactionID: string;
                    outputIndex: number;
                };

                type: string;

                output: {
                    balances: {
                        [color: string]: bigint;
                    };

                    address: string;
                };
            };

            inclusionState: {
                confirmed?: boolean;
                rejected?: boolean;
                conflicting?: boolean;
            };
        }[];
    }[];
}
