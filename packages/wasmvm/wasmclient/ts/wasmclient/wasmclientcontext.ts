// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as cryptolib from "./cryptolib"
import * as wasmlib from "wasmlib"
import * as wc from "./index";
import { error } from "./wasmhost"

export interface IEventHandler {
	CallHandler(topic: string, params: string[]): void;
}

export class WasmClientContext extends wc.WasmClientSandbox  {

	public CurrentChainID(): wasmlib.ScChainID {
		return this.chainID;
	}

	public InitFuncCallContext(): void {
		wasmlib.connectHost(this);
	}

	public InitViewCallContext(hContract: wasmlib.ScHname): wasmlib.ScHname {
		wasmlib.connectHost(this);
		return this.scHname;
	}

	public Register(handler: IEventHandler): error {
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
	public ServiceContractName(contractName: string) {
		this.scHname = wasmlib.ScHname.fromName(contractName);
	}

	public SignRequests(keyPair: cryptolib.KeyPair) {
		this.keyPair = keyPair;
	}

	public Unregister(handler: IEventHandler): void {
		for (let i = 0; i < this.eventHandlers.length; i++) {
			if (this.eventHandlers[i] == handler) {
				let handlers = this.eventHandlers;
				this.eventHandlers = handlers.slice(0, i).concat(handlers.slice(i+1));
				if (this.eventHandlers.length == 0) {
					this.stopEventHandlers();
				}
				return;
			}
		}
	}

	public WaitRequest(reqID: wasmlib.ScRequestID|undefined): error {
		let rID = this.ReqID;
		if (reqID !== undefined) {
			rID = reqID;
		}
		return this.svcClient.WaitUntilRequestProcessed(this.chainID, rID, 60);
	}

	public startEventHandlers(): error {
		let chMsg = make(chan []string, 20);
		this.eventDone = make(chan: bool);
		let err = this.svcClient.SubscribeEvents(chMsg, this.eventDone);
		if (err != null) {
			return err;
		}
		go public(): void {
			for {
				for (let msgSplit = range chMsg) {
					let event = strings.Join(msgSplit, " ");
					fmt.Printf("%this\n", event);
					if (msgSplit[0] == "vmmsg") {
						let msg = strings.Split(msgSplit[3], "|");
						let topic = msg[0];
						let params = msg[1:];
						for (let _,  handler = range this.eventHandlers) {
							handler.CallHandler(topic, params);
						}
					}
				}
			}
		}()
		return null;
	}

	public stopEventHandlers(): void {
		if (len(this.eventHandlers) > 0) {
			this.eventDone <- true;
		}
	}
}
