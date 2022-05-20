package wasmclient

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/mr-tron/base58"
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
	case wasmlib.FnUtilsBase58Decode:
		return Base58Decode(string(args))
	case wasmlib.FnUtilsBase58Encode:
		return []byte(Base58Encode(args))
	case wasmlib.FnUtilsBech32Decode:
		return s.fnUtilsBech32Decode(args)
	case wasmlib.FnUtilsBech32Encode:
		return s.fnUtilsBech32Encode(args)
	}
	panic("implement me")
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
	hContract := s.cvt.IscpHname(req.Contract)
	if hContract != s.scHname {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	params, err := dict.FromBytes(req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	hFunction := s.cvt.IscpHname(req.Function)
	res, err := s.svcClient.CallViewByHname(s.chainID, hContract, hFunction, params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res.Bytes()
}

func (s *WasmClientContext) fnPost(args []byte) []byte {
	req := wasmrequests.NewPostRequestFromBytes(args)
	chainID := s.cvt.IscpChainID(&req.ChainID)
	if !chainID.Equals(s.chainID) {
		s.Err = errors.Errorf("unknown chain id: %s", req.ChainID.String())
		return nil
	}
	hContract := s.cvt.IscpHname(req.Contract)
	if hContract != s.scHname {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	params, err := dict.FromBytes(req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	scAssets := wasmlib.NewScAssets(req.Transfer)
	allowance := s.cvt.IscpAllowance(scAssets)
	hFunction := s.cvt.IscpHname(req.Function)
	s.ReqID, s.Err = s.svcClient.PostRequest(s.chainID, s.scHname, hFunction, params, allowance, s.keyPair)
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
	return s.cvt.ScAddress(addr).Bytes()
}

func (s *WasmClientContext) fnUtilsBech32Encode(args []byte) []byte {
	scAddress := wasmtypes.AddressFromBytes(args)
	addr := s.cvt.IscpAddress(&scAddress)
	return []byte(addr.Bech32(parameters.L1.Protocol.Bech32HRP))
}

/////////////////////////////////////////////////////////////////

func Base58Decode(s string) []byte {
	res, err := base58.Decode(s)
	if err != nil {
		panic("invalid base58 encoding")
	}
	return res
}

func Base58Encode(b []byte) string {
	return base58.Encode(b)
}
