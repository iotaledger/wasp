// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use crate::*;

pub trait IClientService {
	fn callViewByHname(&self, chainID: ScChainID, hContract: ScHname, hFunction: ScHname, args: &[u8]) -> ([u8], error);
	fn postRequest(&self, chainID: ScChainID, hContract: ScHname, hFunction: ScHname, args: &[u8], allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) -> (ScRequestID, error);
	fn subscribeEvents(&self, msg: chan []string, done: chan bool) -> error;
	fn waitUntilRequestProcessed(&self, chainID: ScChainID, reqID: ScRequestID, timeout: time.Duration) -> error;
}

pub struct WasmClientService {
	cvt        : wasmhost.WasmConvertor,
	waspClient : *client.WaspClient,
	eventPort  : String,
	nonce      : u64,
}

impl WasmClientService {
	pub fn new(waspAPI: &str, eventPort: &str) -> *WasmClientService {
		return &WasmClientService{waspClient: client.NewWaspClient(waspAPI), eventPort: eventPort};
	}

	pub fn defaultWasmClientService() -> *WasmClientService {
		return NewWasmClientService("127.0.0.1:9090", "127.0.0.1:5550");
	}

	pub fn callViewByHname(&self, chainID: ScChainID, hContract: ScHname, hFunction: ScHname, args: &[u8]) -> ([u8], error) {
		let iscpChainID = self.cvt.IscpChainID(&chainID);
		let iscpContract = self.cvt.IscpHname(hContract);
		let iscpFunction = self.cvt.IscpHname(hFunction);
		let params,  err = dict.FromBytes(args);
		if err != nil {
			return nil, err;
		}
		let res,  err = self.waspClient.CallViewByHname(iscpChainID, iscpContract, iscpFunction, params);
		if err != nil {
			return nil, err;
		}
		return res.Bytes(), nil;
	}

	pub fn postRequest(&self, chainID: ScChainID, hContract: ScHname, hFunction: ScHname, args: &[u8], allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) -> (reqID: ScRequestID, err: error) {
		let iscpChainID = self.cvt.IscpChainID(&chainID);
		let iscpContract = self.cvt.IscpHname(hContract);
		let iscpFunction = self.cvt.IscpHname(hFunction);
		let params,  err = dict.FromBytes(args);
		if err != nil {
			return reqID, err;
		}
		self.nonce++;
		let req = iscp.NewOffLedgerRequest(iscpChainID, iscpContract, iscpFunction, params, self.nonce);
		let iscpAllowance = self.cvt.IscpAllowance(allowance);
		req.WithAllowance(iscpAllowance);
		let signed = req.Sign(keyPair);
		err = self.waspClient.PostOffLedgerRequest(iscpChainID, signed);
		if err == nil {
			reqID = self.cvt.ScRequestID(signed.ID());
		}
		return reqID, err;
	}

	pub fn subscribeEvents(&self, msg: &Vec<String>, done: chan bool) -> error {
		return subscribe.Subscribe(self.eventPort, msg, done, false, "");
	}

	pub fn waitUntilRequestProcessed(&self, chainID: ScChainID, reqID: ScRequestID, timeout: time.Duration) -> error {
		let iscpChainID = self.cvt.IscpChainID(&chainID);
		let iscpReqID = self.cvt.IscpRequestID(&reqID);
		let _,  err = self.waspClient.WaitUntilRequestProcessed(iscpChainID, *iscpReqID, timeout);
		return err;
	}
}