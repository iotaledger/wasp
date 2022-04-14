package wasmclient

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/pkg/errors"
)

func (s *Service) ExportName(index int32, name string) {
	panic("Service.ExportName")
}

func (s *Service) Sandbox(funcNr int32, args []byte) []byte {
	s.Err = nil
	switch funcNr {
	case wasmlib.FnCall:
		return s.fnCall(args)
	case wasmlib.FnPost:
		return s.fnPost(args)
	}
	panic("implement me")
}

func (s *Service) StateDelete(key []byte) {
	panic("Service.StateDelete")
}

func (s *Service) StateExists(key []byte) bool {
	panic("Service.StateExists")
}

func (s *Service) StateGet(key []byte) []byte {
	panic("Service.StateGet")
}

func (s *Service) StateSet(key, value []byte) {
	panic("Service.StateSet")
}

/////////////////////////////////////////////////////////////////

func (s *Service) fnCall(args []byte) []byte {
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
	hFunction := iscp.Hname(req.Function)
	res, err := s.waspClient.CallViewByHname(s.chainID, hContract, hFunction, params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res.Bytes()
}

func (s *Service) fnPost(args []byte) []byte {
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
	s.Req = s.postRequestOffLedger(hFunction, params, allowance, s.keyPair)
	return nil
}
