// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {WasmClientSandbox} from './wasmclientsandbox';
import {WasmClientService} from './wasmclientservice';
import {WasmClientEvents} from "./wasmclientevents";

export class WasmClientContext extends WasmClientSandbox implements wasmlib.ScFuncCallContext {

    public constructor(svcClient: WasmClientService, scName: string) {
        super(svcClient, scName);
    }

    public currentKeyPair(): isc.KeyPair | null {
        return this.keyPair;
    }

    public currentSvcClient(): WasmClientService {
        return this.svcClient;
    }

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    public initFuncCallContext(): void {
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public initViewCallContext(_hContract: wasmlib.ScHname): wasmlib.ScHname {
        return this.scHname;
    }

    public register(handler: wasmlib.IEventHandlers): isc.Error {
        return this.svcClient.subscribeEvents(new WasmClientEvents(
            this.svcClient.currentChainID(),
            this.scHname,
            handler
        ));
    }

    public signRequests(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;
    }

    public unregister(eventsID: u32): void {
        this.svcClient.unsubscribeEvents(eventsID);
    }

    public waitRequest(): void {
        this.waitRequestID(this.ReqID);
    }

    public waitRequestID(reqID: wasmlib.ScRequestID): void {
        this.Err = this.svcClient.waitUntilRequestProcessed(reqID, 60);
    }
}
