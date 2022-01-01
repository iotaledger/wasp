// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {IKeyPair} from "./crypto";

export class ClientFunc {
    protected svc: wasmclient.Service;
    private keyPair: IKeyPair;
    private xfer: wasmclient.Transfer = new wasmclient.Transfer();

    constructor(svc: wasmclient.Service) {
        this.svc = svc;
    }

    // Sends a request to the smart contract service
    // You can wait for the request to complete by using the returned RequestID
    // as parameter to Service.waitRequest()
    public async post(hFuncName: wasmclient.Hname, args: wasmclient.Arguments | null): Promise<wasmclient.RequestID> {
        if (args == null) {
            args = new wasmclient.Arguments();
        }
        if (this.keyPair == null) {
            this.keyPair = this.svc.keyPair;
        }
        return await this.svc.postRequest(hFuncName, args, this.xfer, this.keyPair);
    }

    // Optionally overrides the default keypair from the service
    public sign(keyPair: IKeyPair): void {
        this.keyPair = keyPair;
    }

    // Optionally indicates which tokens to transfer as part of the request
    // The tokens are presumed to be available in the signing account on the chain
   public transfer(xfer: wasmclient.Transfer): void {
        this.xfer = xfer;
    }
}
