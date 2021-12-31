// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"

export class ClientView {
    private svc: wasmclient.Service;
    private res: wasmclient.Results;

    constructor(svc: wasmclient.Service) {
        this.svc = svc;
    }

    call(viewName: string, args: wasmclient.Arguments | null): void {
        if (args == null) {
            args = new wasmclient.Arguments();
        }
        this.res = this.svc.callView(viewName, args);
    }

    results(): wasmclient.Results {
        return this.res;
    }
}
