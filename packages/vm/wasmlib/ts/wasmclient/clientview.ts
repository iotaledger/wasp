// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"

export class ClientView {
    private svc: wasmclient.Service;

    constructor(svc: wasmclient.Service) {
        this.svc = svc;
    }

    protected async callView(viewName: string, args: wasmclient.Arguments | null): Promise<wasmclient.Results> {
        if (args == null) {
            args = new wasmclient.Arguments();
        }
        return await this.svc.callView(viewName, args);
    }
}
