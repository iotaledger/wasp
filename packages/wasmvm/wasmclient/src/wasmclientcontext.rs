// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub trait IEventHandler {
	fn call_handler(&self, topic: &str, params: &Vec<String>);
}

pub struct WasmClientContext {
	chainID: ScChainID,
	Err: error,
	eventDone: bool,
	eventHandlers: Vec<dyn IEventHandler>,
	keyPair: * cryptolib.KeyPair,
	ReqID: ScRequestID,
	scName: String,
	scHname: ScHname,
	svcClient: IClientService,
}

impl WasmClientContext {
	pub fn new(svcClient: IClientService, chainID: ScChainID, scName: &str) -> WasmClientContext {
		WasmClientContext{
			svcClient : svcClient,
			scName : scName.to_string(),
			scHname : ScHname::new(scName),
			chainID : chainID,
			Err: (),
			eventDone: false,
			eventHandlers: (),
			keyPair: (),
			ReqID: ()
		}
	}

	pub fn CurrentChainID(&self) -> ScChainID {
		return self.chainID;
	}

	pub fn InitFuncCallContext(&self) {
		connect_host(self);
	}

	pub fn InitViewCallContext(&self, hContract: ScHname) -> ScHname {
		connect_host(self);
		return self.scHname;
	}

	pub fn Register(&self, handler: dyn IEventHandler) -> error {
		for h in self.eventHandlers {
			if h == handler {
				return nil;
			}
		}
		self.eventHandlers = append(self.eventHandlers, handler);
		if len(self.eventHandlers) > 1 {
			return nil;
		}
		return self.startEventHandlers();
	}

	// overrides default contract name
	pub fn ServiceContractName(&self, contractName: &str) {
		self.scHname = NewScHname(contractName);
	}

	pub fn SignRequests(&self, keyPair: *cryptolib.KeyPair) {
		self.keyPair = keyPair;
	}

	pub fn Unregister(&self, handler: dyn IEventHandler) {
		for h in self.eventHandlers {
			if h == handler {
				self.eventHandlers = append(self.eventHandlers[:i], self.eventHandlers[i+1:]...);
				if len(self.eventHandlers) == 0 {
					self.stopEventHandlers();
				}
				return;
			}
		}
	}

	pub fn WaitRequest(&self, reqID: ...ScRequestID) -> error {
		let requestID = self.ReqID;
		if len(reqID) == 1 {
			requestID = reqID[0];
		}
		return self.svcClient.WaitUntilRequestProcessed(self.chainID, requestID, 1*time.Minute);
	}

	pub fn startEventHandlers(&self) -> error {
		let chMsg = make(chan []string, 20);
		self.eventDone = make(chan: bool);
		let err = self.svcClient.SubscribeEvents(chMsg, self.eventDone);
		if err != nil {
			return err;
		}
		go pub fn() {
			for {
				for let msgSplit = range chMsg {
					let event = strings.Join(msgSplit, " ");
					fmt.Printf("%self\n", event);
					if msgSplit[0] == "vmmsg" {
						let msg = strings.Split(msgSplit[3], "|");
						let topic = msg[0];
						let params = msg[1:];
						for let _,  handler = range self.eventHandlers {
							handler.CallHandler(topic, params);
						}
					}
				}
			}
		}()
		return nil;
	}

	pub fn stopEventHandlers(&self) {
		if len(self.eventHandlers) > 0 {
			self.eventDone <- true;
		}
	}
}
