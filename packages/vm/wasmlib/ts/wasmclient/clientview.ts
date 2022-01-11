// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"

export class ClientView {
    private svc: wasmclient.Service;

    constructor(svc: wasmclient.Service) {
        this.svc = svc;
    }

    protected async callView(viewName: string, args: wasmclient.Arguments | null, res: wasmclient.Results): Promise<void> {
        if (args == null) {
            args = new wasmclient.Arguments();
        }
        await this.svc.callView(viewName, args, res);
    }
}
