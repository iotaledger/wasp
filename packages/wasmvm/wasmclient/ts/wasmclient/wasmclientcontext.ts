// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wc from "./index";

export interface IEventHandler {
	CallHandler(topic: string, params: string[]): void;
}

export class WasmClientContext extends wc.WasmClientSandbox  {
	chainID       : wasmlib.ScChainID;
	Err           : error;
	eventDone     : bool;
	eventHandlers : IEventHandler[];
	keyPair       : cryptolib.KeyPair;
	ReqID         : wasmlib.ScRequestID;
	scName        : string;
	scHname       : wasmlib.ScHname;
	svcClient     : wc.IClientService;

	public constructor(svcClient: wc.IClientService, chainID: wasmlib.ScChainID, scName: string) {
		super();
		this.svcClient = svcClient;
		this.scName = scName;
		this.scHname = wasmlib.ScHname.fromName(scName);
		this.chainID = chainID;
	}

	public CurrentChainID(): wasmlib.ScChainID {
		return this.chainID;
	}

	public InitFuncCallContext(): void {
		wasmhost.Connect(this);
	}

	public InitViewCallContext(hContract: wasmlib.ScHname): wasmlib.ScHname {
		wasmhost.Connect(this);
		return this.scHname;
	}

	public Register(handler: IEventHandler): error {
		for (let i = 0; i < this.eventHandlers.length; i++) {
			if (this.eventHandlers[i] == handler) {
				return nil;
			}
		}
		this.eventHandlers.push(handler);
		if (this.eventHandlers.length > 1) {
			return nil;
		}
		return this.startEventHandlers();
	}

	// overrides default contract name
	public ServiceContractName(contractName: string) {
		this.scHname = new wasmlib.ScHname(contractName);
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

	public WaitRequest(reqID: ...wasmlib.ScRequestID): error {
		let requestID = this.ReqID;
		if (len(reqID) == 1) {
			requestID = reqID[0];
		}
		return this.svcClient.WaitUntilRequestProcessed(this.chainID, requestID, 1*time.Minute);
	}

	public startEventHandlers(): error {
		let chMsg = make(chan []string, 20);
		this.eventDone = make(chan: bool);
		let err = this.svcClient.SubscribeEvents(chMsg, this.eventDone);
		if (err != nil) {
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
		return nil;
	}

	public stopEventHandlers(): void {
		if (len(this.eventHandlers) > 0) {
			this.eventDone <- true;
		}
	}
}
