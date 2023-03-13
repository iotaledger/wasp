// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as coreaccounts from 'wasmlib/coreaccounts';
import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import {WebSocket} from 'ws';
import {WasmClientContext} from './wasmclientcontext';
import {WasmClientEvents} from "./wasmclientevents";

export class WasmClientService {
    private chainID: wasmlib.ScChainID;
    //TODO do we need to lock a mutex here?
    private eventHandlers: WasmClientEvents[] = [];
    //TODO do we need to lock a mutex here?
    private nonces = new Map<Uint8Array, u64>();
    private waspAPI: string;
    private ws: WebSocket;

    public constructor(waspAPI: string, chainID: string) {
        const err = isc.setSandboxWrappers(chainID);
        if (err != null) {
            panic(err);
        }
        this.waspAPI = waspAPI;
        this.chainID = wasmlib.chainIDFromString(chainID);
        const url = waspAPI.replace('http:', 'ws:') + '/ws';
        this.ws = new WebSocket(url, {
            perMessageDeflate: false
        });
    }

    public callViewByHname(hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, isc.Error] {
        const callViewRequest: isc.APICallViewRequest = {
            contractHName: hContract.toString(),
            functionHName: hFunction.toString(),
            chainId: this.chainID.toString(),
            arguments: isc.Codec.jsonEncode(args),
        };

        const url = this.waspAPI + '/requests/callview';
        const client = new isc.SyncRequestClient();
        client.addHeader('Content-Type', 'application/json');
        try {
            const resp = client.post<isc.APICallViewRequest, isc.JsonResp>(url, callViewRequest);
            const result = isc.Codec.jsonDecode(resp);
            return [result, null];
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return [new Uint8Array(0), message];
        }
    }

    public currentChainID(): wasmlib.ScChainID {
        return this.chainID;
    }

    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair): [wasmlib.ScRequestID, isc.Error] {
        const [nonce, err] = this.cachedNonce(keyPair);
        if (err != null) {
            return [new wasmlib.ScRequestID(), err];
        }
        const req = new isc.OffLedgerRequest(chainID, hContract, hFunction, args, nonce);
        req.withAllowance(allowance);
        const signed = req.sign(keyPair);
        const reqID = signed.ID();

        const offLedgerRequest: isc.APIOffLedgerRequest = {
            chainId: chainID.toString(),
            request: wasmlib.hexEncode(signed.bytes()),
        };

        const url = this.waspAPI + '/requests/offledger';
        const client = new isc.SyncRequestClient();
        client.addHeader('Content-Type', 'application/json');
        try {
            client.post(url, offLedgerRequest);
            return [reqID, null];
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return [reqID, message];
        }
    }

    public subscribeEvents(eventHandler: WasmClientEvents): isc.Error {
        this.eventHandlers.push(eventHandler);
        if (this.eventHandlers.length != 1) {
            return null;
        }
        return WasmClientEvents.startEventLoop(this.ws, this.eventHandlers)
    }

    public unsubscribeEvents(eventsID: u32): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i].handler.id() == eventsID) {
                this.eventHandlers.splice(i, 1);
            }
        }
        if (this.eventHandlers.length == 0) {
            // stop event loop
            this.ws.close();
        }
    }

    public waitUntilRequestProcessed(reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        //TODO Timeout of the wait can be set with `/wait?timeoutSeconds=`. Max seconds are 60secs.
        const url = this.waspAPI + '/chains/' + this.chainID.toString() + '/requests/' + reqID.toString() + '/wait';
        try {
            new isc.SyncRequestClient().get(url);
            return null;
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return message;
        }
    }

    private cachedNonce(keyPair: isc.KeyPair): [u64, isc.Error] {
        const key = keyPair.publicKey;
        let nonce = this.nonces.get(key);
        if (nonce === undefined) {
            const agent = wasmlib.ScAgentID.fromAddress(keyPair.address());
            const ctx = new WasmClientContext(this, coreaccounts.ScName);
            const n = coreaccounts.ScFuncs.getAccountNonce(ctx);
            n.params.agentID().setValue(agent);
            n.func.call();
            if (ctx.Err != null) {
                return [0n, ctx.Err];
            }
            nonce = n.results.accountNonce().value();
        }
        nonce++;
        this.nonces.set(key, nonce);
        return [nonce, null];
    }
}
