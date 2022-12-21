// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import * as coreaccounts from 'wasmlib/coreaccounts';
import {WasmClientSandbox} from './wasmclientsandbox';
import {IClientService} from "./wasmclientservice";

export class WasmClientContext extends WasmClientSandbox implements wasmlib.ScFuncCallContext {
    private eventDone: bool = false;
    private eventHandlers: wasmlib.IEventHandlers[] = [];
    private eventReceived: bool = false;
    private webSocket!: WebSocket;

    public constructor(svcClient: IClientService, chain: string, scName: string) {
        super(svcClient, chain, scName);
        //this.connectWebSocket();
    }

    private connectWebSocket(): void {
        const webSocketUrl = "ws://127.0.0.1:9090/chain/" + this.chainID.toString() + "/ws";
        // eslint-disable-next-line no-console
        console.log(`Connecting to Websocket => ${webSocketUrl}`);
        this.webSocket = new WebSocket(webSocketUrl);
        this.webSocket.addEventListener('error', console.log);
        this.webSocket.addEventListener('message', (x) => {
            console.log(x);
            this.handleIncomingMessage(x);
        });
        this.webSocket.addEventListener('close', () => setTimeout(this.connectWebSocket.bind(this), 1000));
    }

    private handleIncomingMessage(message: MessageEvent<string>): void {
        // expect vmmsg <chain ID> <contract hname> contract.event|param1|param2|...
        const msg = message.data.toString().split(' ');
        if (msg.length != 4 || msg[0] != 'vmmsg') {
            return;
        }
        this.eventDone = true;
    }

    public currentChainID(): wasmlib.ScChainID {
        return this.chainID;
    }

    public initFuncCallContext(): void {
        wasmlib.connectHost(this);
    }

    public initViewCallContext(_hContract: wasmlib.ScHname): wasmlib.ScHname {
        wasmlib.connectHost(this);
        return this.scHname;
    }

    public register(handler: wasmlib.IEventHandlers): isc.Error {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] == handler) {
                return null;
            }
        }
        this.eventHandlers.push(handler);
        if (this.eventHandlers.length > 1) {
            return null;
        }
        return this.startEventHandlers();
    }

    // overrides default contract name
    public serviceContractName(contractName: string) {
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(contractName));
    }

    public signRequests(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;

        // get last used nonce from accounts core contract
        const agent = wasmlib.ScAgentID.fromAddress(keyPair.address());
        const ctx = new WasmClientContext(this.svcClient, this.chainID.toString(), coreaccounts.ScName);
        const n = coreaccounts.ScFuncs.getAccountNonce(ctx);
        n.params.agentID().setValue(agent);
        n.func.call();
        this.nonce = n.results.accountNonce().value();
    }

    public unregister(handler: wasmlib.IEventHandlers): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] == handler) {
                const handlers = this.eventHandlers;
                this.eventHandlers = handlers.slice(0, i).concat(handlers.slice(i + 1));
                if (this.eventHandlers.length == 0) {
                    this.stopEventHandlers();
                }
                return;
            }
        }
    }

    public waitEvent(callback: () => void): void {
        this.Err = null;
        let counter = 0;
        const handle = setInterval(() => {
            if (this.eventReceived) {
                console.log("Event received");
                clearInterval(handle);
                callback();
                return;
            }

            if (counter >= 100) {
                clearInterval(handle);
                this.Err = "event wait timeout";
                callback();
                return;
            }

            counter++;
        }, 100);
    }

    public waitRequest(): void {
        this.waitRequestID(this.ReqID);
    }

    public waitRequestID(reqID: wasmlib.ScRequestID): void {
        this.Err = this.svcClient.waitUntilRequestProcessed(this.chainID, reqID, 60);
    }

    public startEventHandlers(): isc.Error {
        //TODO
        // let chMsg = make(chan []string, 20);
        // this.eventDone = make(chan: bool);
        // let err = this.svcClient.SubscribeEvents(chMsg, this.eventDone);
        // if (err != null) {
        // 	return err;
        // }
        // go public(): void {
        // 	for {
        // 		for (let msgSplit = range chMsg) {
        // 			let event = strings.Join(msgSplit, ' ');
        // 			fmt.Printf('%this\n', event);
        // 			if (msgSplit[0] == 'vmmsg') {
        // 				let msg = strings.Split(msgSplit[3], '|');
        // 				let topic = msg[0];
        // 				let params = msg[1:];
        // 				for (let _,  handler = range this.eventHandlers) {
        // 					handler.CallHandler(topic, params);
        // 				}
        // 			}
        // 		}
        // 	}
        // }()
        return null;
    }

    public stopEventHandlers(): void {
        //TODO
        // if (this.eventHandlers.length > 0) {
        // 	this.eventDone <- true;
        // }
    }
}
