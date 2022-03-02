package wasmclient

import (
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
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
	contract := uint32(req.Contract)
	if contract != uint32(s.scHname) {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	params, err := dict.FromBytes(req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	reqArgs := requestargs.New()
	reqArgs.AddEncodeSimpleMany(params)
	res, err := s.waspClient.CallView(s.chainID, s.scHname, "0x"+req.Function.String(), params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res.Bytes()
}

func (s *Service) fnPost(args []byte) []byte {
	req := wasmrequests.NewPostRequestFromBytes(args)
	if req.ChainID != s.ChainID() {
		s.Err = errors.Errorf("unknown chain id: %s", req.ChainID.String())
		return nil
	}
	contract := uint32(req.Contract)
	if contract != uint32(s.scHname) {
		s.Err = errors.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}
	params, err := dict.FromBytes(req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	bal, err := colored.BalancesFromBytes(req.Transfer)
	if err != nil {
		s.Err = err
		return nil
	}
	//if len(bal) == 0 && !s.offLedger {
	//	bal.Add(colored.Color{}, 1)
	//}
	reqArgs := requestargs.New()
	reqArgs.AddEncodeSimpleMany(params)
	s.Req = s.postRequestOffLedger(uint32(req.Function), reqArgs, bal, s.keyPair)
	return nil
}
