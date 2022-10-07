// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"github.com/pkg/errors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

func (s *WasmClientContext) ExportName(index int32, name string) {
	_ = index
	_ = name
	panic("WasmClientContext.ExportName")
}

func (s *WasmClientContext) Sandbox(funcNr int32, args []byte) []byte {
	s.Err = nil
	switch funcNr {
	case wasmlib.FnCall:
		s.eventReceived = false
		return s.fnCall(args)
	case wasmlib.FnPost:
		s.eventReceived = false
		return s.fnPost(args)
	case wasmlib.FnUtilsBech32Decode:
		return s.fnUtilsBech32Decode(args)
	case wasmlib.FnUtilsBech32Encode:
		return s.fnUtilsBech32Encode(args)
	case wasmlib.FnUtilsHashName:
		return s.fnUtilsHashName(args)
	}
	panic("implement WasmClientContext.Sandbox")
}

func (s *WasmClientContext) StateDelete(key []byte) {
	_ = key
	panic("WasmClientContext.StateDelete")
}

func (s *WasmClientContext) StateExists(key []byte) bool {
	_ = key
	panic("WasmClientContext.StateExists")
}

func (s *WasmClientContext) StateGet(key []byte) []byte {
	_ = key
	panic("WasmClientContext.StateGet")
}

func (s *WasmClientContext) StateSet(key, value []byte) {
	_ = key
	_ = value
	panic("WasmClientContext.StateSet")
}

/////////////////////////////////////////////////////////////////

func (s *WasmClientContext) fnCall(args []byte) []byte {
	req := wasmrequests.NewCallRequestFromBytes(args)
	if req.Contract != s.scHname {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	res, err := s.svcClient.CallViewByHname(s.chainID, req.Contract, req.Function, req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res
}

func (s *WasmClientContext) fnPost(args []byte) []byte {
	req := wasmrequests.NewPostRequestFromBytes(args)
	if req.ChainID != s.chainID {
		s.Err = errors.Errorf("unknown chain id: %s", req.ChainID.String())
		return nil
	}
	if req.Contract != s.scHname {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	scAssets := wasmlib.NewScAssets(req.Transfer)
	s.ReqID, s.Err = s.svcClient.PostRequest(s.chainID, req.Contract, req.Function, req.Params, scAssets, s.keyPair)
	return nil
}

func (s *WasmClientContext) fnUtilsBech32Decode(args []byte) []byte {
	hrp, addr, err := iotago.ParseBech32(string(args))
	if err != nil {
		s.Err = err
		return nil
	}
	if hrp != parameters.L1().Protocol.Bech32HRP {
		s.Err = errors.Errorf("Invalid protocol prefix: %s", string(hrp))
		return nil
	}
	var cvt wasmhost.WasmConvertor
	return cvt.ScAddress(addr).Bytes()
}

func (s *WasmClientContext) fnUtilsBech32Encode(args []byte) []byte {
	var cvt wasmhost.WasmConvertor
	scAddress := wasmtypes.AddressFromBytes(args)
	addr := cvt.IscAddress(&scAddress)
	return []byte(addr.Bech32(parameters.L1().Protocol.Bech32HRP))
}

func (s *WasmClientContext) fnUtilsHashName(args []byte) []byte {
	return isc.Hn(string(args)).Bytes()
}
