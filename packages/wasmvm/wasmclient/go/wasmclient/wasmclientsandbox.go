// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/pkg/errors"
)

func (s *WasmClientContext) ExportName(index int32, name string) {
	panic("WasmClientContext.ExportName")
}

func (s *WasmClientContext) Sandbox(funcNr int32, args []byte) []byte {
	s.Err = nil
	switch funcNr {
	case wasmlib.FnCall:
		return s.fnCall(args)
	case wasmlib.FnPost:
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
	panic("WasmClientContext.StateDelete")
}

func (s *WasmClientContext) StateExists(key []byte) bool {
	panic("WasmClientContext.StateExists")
}

func (s *WasmClientContext) StateGet(key []byte) []byte {
	panic("WasmClientContext.StateGet")
}

func (s *WasmClientContext) StateSet(key, value []byte) {
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
	if hrp != parameters.L1.Protocol.Bech32HRP {
		s.Err = errors.Errorf("Invalid protocol prefix: %s", string(hrp))
		return nil
	}
	var cvt wasmhost.WasmConvertor
	return cvt.ScAddress(addr).Bytes()
}

func (s *WasmClientContext) fnUtilsBech32Encode(args []byte) []byte {
	var cvt wasmhost.WasmConvertor
	scAddress := wasmtypes.AddressFromBytes(args)
	addr := cvt.IscpAddress(&scAddress)
	return []byte(addr.Bech32(parameters.L1.Protocol.Bech32HRP))
}

func (s *WasmClientContext) fnUtilsHashName(args []byte) []byte {
	var utils iscp.Utils
	return codec.EncodeHname(utils.Hashing().Hname(string(args)))
}
