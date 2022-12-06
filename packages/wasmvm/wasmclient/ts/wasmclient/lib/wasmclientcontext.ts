// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Ed25519Address } from '@iota/iota.js';
import * as coreaccounts from 'wasmlib/coreaccounts';
import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import { WasmClientSandbox } from './wasmclientsandbox';

export class WasmClientContext extends WasmClientSandbox {

    public chainID(): wasmlib.ScChainID {
        return this.chID;
    }

    public initFuncCallContext(): void {
        wasmlib.connectHost(this);
    }

    public initViewCallContext(hContract: wasmlib.ScHname): wasmlib.ScHname {
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
        const iscAddr = new Ed25519Address(keyPair.publicKey).toAddress();
        //TODO iscAddr convert to ScAddress
        const addr = wasmlib.addressFromBytes(wasmlib.bytesFromUint8Array(iscAddr));
        const agent = wasmlib.ScAgentID.fromAddress(addr);
        const ctx = new WasmClientContext(this.svcClient, this.chID, coreaccounts.ScName);
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

    public waitRequest(reqID: wasmlib.ScRequestID | undefined): isc.Error {
        let rID = this.ReqID;
        if (reqID !== undefined) {
            rID = reqID;
        }
        return this.svcClient.waitUntilRequestProcessed(this.chID, rID, 60);
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
