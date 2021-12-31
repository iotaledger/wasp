// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"

export class ServiceClient {
    waspClient: wasmclient.WaspClient;
    eventPort: string;

    constructor(waspAPI: string, eventPort: string) {
        this.waspClient = new wasmclient.WaspClient(waspAPI);
        this.eventPort = eventPort;
    }

    static default(): ServiceClient {
        //TODO use TCP instead of websocket for event listener?
        return new ServiceClient("127.0.0.1:9090", "127.0.0.1:9090"); // "127.0.0.1:5550");
    }
}
