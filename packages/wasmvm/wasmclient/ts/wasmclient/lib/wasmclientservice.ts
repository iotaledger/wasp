// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as coreaccounts from 'wasmlib/coreaccounts';
import * as iscclient from './iscclient';
import * as wasmlib from 'wasmlib';
import {WebSocket} from 'ws';
import {WasmClientContext} from './wasmclientcontext';
import {WasmClientEvents} from './wasmclientevents';

class ChainInfoResponse {
    chainID = '';
}

export class WasmClientService {
    private chainID: wasmlib.ScChainID;
    //TODO do we need to lock a mutex here?
    private eventHandlers: WasmClientEvents[] = [];
    //TODO do we need to lock a mutex here?
    private nonces = new Map<Uint8Array, u64>();
    private waspAPI: string;
    private ws: WebSocket | null = null;

    public constructor(waspAPI: string) {
        this.waspAPI = waspAPI;
        this.chainID = wasmlib.chainIDFromBytes(null);
    }

    public callViewByHname(hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, iscclient.Error] {
        const url = this.waspAPI + '/v1/chains/' + this.chainID.toString() + '/callview';
        const callViewRequest: iscclient.APICallViewRequest = {
            contractHName: hContract.toString(),
            functionHName: hFunction.toString(),
            arguments: iscclient.Codec.jsonEncode(args),
        };
        try {
            const client = new iscclient.SyncRequestClient();
            client.addHeader('Content-Type', 'application/json');
            const resp = client.post<iscclient.APICallViewRequest, iscclient.JsonResp>(url, callViewRequest);
            const result = iscclient.Codec.jsonDecode(resp);
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

    public isHealthy(): bool {
        const url = this.waspAPI + '/health';
        try {
            new iscclient.SyncRequestClient().get(url);
            return true;
        } catch (error) {
            return false;
        }
    }

    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: iscclient.KeyPair): [wasmlib.ScRequestID, iscclient.Error] {
        const [nonce, err] = this.cachedNonce(keyPair);
        if (err != null) {
            return [new wasmlib.ScRequestID(), err];
        }
        const req = new iscclient.OffLedgerRequest(chainID, hContract, hFunction, args, nonce);
        req.withAllowance(allowance);
        req.sign(keyPair);
        const reqID = req.ID();

        const url = this.waspAPI + '/v1/requests/offledger';
        const offLedgerRequest: iscclient.APIOffLedgerRequest = {
            chainId: chainID.toString(),
            request: wasmlib.hexEncode(req.bytes()),
        };
        try {
            const client = new iscclient.SyncRequestClient();
            client.addHeader('Content-Type', 'application/json');
            client.post(url, offLedgerRequest);
            return [reqID, null];
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return [reqID, message];
        }
    }

    public setCurrentChainID(chainID: string): iscclient.Error {
        const err = iscclient.setSandboxWrappers(chainID);
        if (err != null) {
            return err;
        }
        this.chainID = wasmlib.chainIDFromString(chainID);
        return null;
    }

    public setDefaultChainID(): iscclient.Error {
        const url = this.waspAPI + '/v1/chains';
        try {
            const client = new iscclient.SyncRequestClient();
            client.addHeader('Content-Type', 'application/json');
            const chains = client.get<ChainInfoResponse[]>(url);
            if (chains.length != 1) {
                return 'expected a single chain for default chain ID';
            }
            const chainID = chains[0].chainID;
            console.log('default chain ID: ' + chainID)
            return this.setCurrentChainID(chainID);
        } catch (error) {
            if (error instanceof Error) return error.message;
            return String(error);
        }
    }

    public subscribeEvents(eventHandler: WasmClientEvents): iscclient.Error {
        this.eventHandlers.push(eventHandler);
        if (this.eventHandlers.length != 1) {
            return null;
        }
        const url = this.waspAPI.replace('http:', 'ws:') + '/v1/ws';
        this.ws = new WebSocket(url, {
            perMessageDeflate: false
        });
        return WasmClientEvents.startEventLoop(this.ws, this.eventHandlers)
    }

    public unsubscribeEvents(eventsID: u32): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i].handler.id() == eventsID) {
                this.eventHandlers.splice(i, 1);
            }
        }
        if (this.eventHandlers.length == 0 && this.ws != null) {
            // stop event loop
            this.ws.close();
            this.ws = null;
        }
    }

    public waitUntilRequestProcessed(reqID: wasmlib.ScRequestID, timeout: u32): iscclient.Error {
        //TODO Timeout of the wait can be set with `/wait?timeoutSeconds=`. Max seconds are 60secs.
        const url = this.waspAPI + '/v1/chains/' + this.chainID.toString() + '/requests/' + reqID.toString() + '/wait';
        try {
            const client = new iscclient.SyncRequestClient();
            client.get(url);
            return null;
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return message;
        }
    }

    private cachedNonce(keyPair: iscclient.KeyPair): [u64, iscclient.Error] {
        const key = keyPair.publicKey;
        let nonce = this.nonces.get(key);
        if (nonce !== undefined) {
            this.nonces.set(key, nonce + 1n);
            return [nonce, null];
        }

        const agent = wasmlib.ScAgentID.fromAddress(keyPair.address());
        const ctx = new WasmClientContext(this, coreaccounts.ScName);
        const n = coreaccounts.ScFuncs.getAccountNonce(ctx);
        n.params.agentID().setValue(agent);
        n.func.call();
        if (ctx.Err != null) {
            return [0n, ctx.Err];
        }
        nonce = n.results.accountNonce().value();
        this.nonces.set(key, nonce + 1n);
        return [nonce, null];
    }
}
