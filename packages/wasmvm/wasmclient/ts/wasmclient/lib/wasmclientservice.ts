// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as coreaccounts from 'wasmlib/coreaccounts';
import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import {RawData, WebSocket} from 'ws';
import {WasmClientContext} from './wasmclientcontext';

export class ContractEvent {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    data: string;

    public constructor(chainID: string, contractID: string, data: string) {
        this.chainID = wasmlib.chainIDFromString(chainID);
        this.contractID = wasmlib.hnameFromString(contractID);
        this.data = data;
    }
}

type ClientCallBack = (event: ContractEvent) => void;

export class WasmClientService {
    private callbacks: ClientCallBack[] = [];
    private chainID: wasmlib.ScChainID;
    private nonces = new Map<Uint8Array, u64>();
    private subscribers: WasmClientContext[] = [];
    private waspAPI: string;
    private ws: WebSocket;

    public constructor(waspAPI: string, chainID: string) {
        const err = isc.setSandboxWrappers(chainID);
        if (err != null) {
            panic(err);
        }
        this.waspAPI = waspAPI;
        this.chainID = wasmlib.chainIDFromString(chainID);
        const eventPort = waspAPI.replace('http:', 'ws:') + '/ws';
        this.ws = new WebSocket(eventPort, {
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

    public subscribeEvents(who: WasmClientContext, callback: ClientCallBack): isc.Error {
        // eslint-disable-next-line @typescript-eslint/no-this-alias
        this.subscribers.push(who);
        this.callbacks.push(callback);
        if (this.subscribers.length == 1) {
            this.ws.on('open', () => {
                this.eventSubscribe('chains');
                this.eventSubscribe('block_events');
            });
            this.ws.on('error', (err) => {
                // callback(['error', err.toString()]);
            });
            this.ws.on('message', (data) => this.eventLoop(data));
        }
        return null;
    }

    public unsubscribeEvents(who: WasmClientContext): void {
        for (let i = 0; i < this.subscribers.length; i++) {
            if (this.subscribers[i] === who) {
                this.subscribers.splice(i, 1);
                this.callbacks.splice(i, 1);
                if (this.subscribers.length == 0) {
                    this.ws.close();
                }
                return;
            }
        }
    }

    public waitUntilRequestProcessed(reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        //TODO Timeout of the wait can be set with `/wait?timeoutSeconds=`. Max seconds are 60secs.
        const url = this.waspAPI + '/chains/' + this.chainID.toString() + '/requests/' + reqID.toString() + '/wait';
        new isc.SyncRequestClient().get(url);
        return null;
    }

    private cachedNonce(keyPair: isc.KeyPair): [u64, isc.Error] {
        //TODO do we need to lock a nonceLock mutex here?
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

    private eventLoop(data: RawData) {
        let msg: any;
        try {
            const json = data.toString();
            // console.log(json);
            msg = JSON.parse(json);
            if (!msg.kind) {
                // filter out subscribe responses
                return;
            }
            console.log(msg);
        } catch (ex) {
            console.log(`Failed to parse expected JSON message: ${data} ${ex}`);
            return;
        }

        const items: string[] = msg.payload;
        for (const item of items) {
            const parts = item.split(': ');
            const event = new ContractEvent(msg.chainID, parts[0], parts[1]);
            for (const callback of this.callbacks) {
                callback(event);
            }
        }
    }

    private eventSubscribe(topic: string) {
        const msg = {
            command: 'subscribe',
            topic: topic,
        };
        const rawMsg = JSON.stringify(msg);
        this.ws.send(rawMsg);
    }
}
