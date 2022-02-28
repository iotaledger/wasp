// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Buffer} from './buffer';

export interface IConfiguration {
    seed: Buffer | null;
    waspWebSocketUrl: string;
    waspApiUrl: string;
    goShimmerApiUrl: string;
}

export class Configuration implements IConfiguration {
    seed: Buffer | null = null;
    waspWebSocketUrl: string = '';
    waspApiUrl: string = '';
    goShimmerApiUrl: string = '';
    chainId: string = '';

    constructor(configuration: IConfiguration) {
        if (!configuration) throw new Error("Configuration not defined");

        this.seed = configuration.seed;
        this.waspWebSocketUrl = configuration.waspWebSocketUrl;
        this.waspApiUrl = configuration.waspApiUrl;
        this.goShimmerApiUrl = configuration.goShimmerApiUrl;
    }

    public toString = (): string => {
        return `Configuration : { seed: ${this.seed}, waspWebSocketUrl : ${this.waspWebSocketUrl}, waspApiUrl : ${this.waspApiUrl}, goShimmerApiUrl : ${this.goShimmerApiUrl}, chainId : ${this.chainId}}`;
    };
}
