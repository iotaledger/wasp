package wasmsolo

import (
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/pkg/errors"
)

type SoloClientService struct {
	ctx   *SoloContext
	msg   chan []string
	nonce uint64
}

var _ wasmclient.IClientService = new(SoloClientService)

// NewSoloClientService creates a new SoloClientService
// Normally we reset the subscribers, assuming a new test.
// To prevent this when testing with multiple SoloClients,
// use the optional extra flag to indicate the extra clients.
func NewSoloClientService(ctx *SoloContext, extra ...bool) *SoloClientService {
	s := &SoloClientService{ctx: ctx}
	if len(extra) != 1 || !extra[0] {
		wasmhost.EventSubscribers = nil
	}
	wasmhost.EventSubscribers = append(wasmhost.EventSubscribers, func(msg string) {
		s.Event(msg)
	})
	return s
}

func (s *SoloClientService) CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscpChainID := s.ctx.Cvt.IscpChainID(&chainID)
	iscpContract := s.ctx.Cvt.IscpHname(hContract)
	iscpFunction := s.ctx.Cvt.IscpHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	if !iscpChainID.Equals(s.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.CallViewByHname chain ID mismatch")
	}
	res, err := s.ctx.Chain.CallViewByHname(iscpContract, iscpFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (s *SoloClientService) Event(msg string) {
	msg = "vmmsg " + s.ctx.CurrentChainID().String() + " 0 " + msg
	s.msg <- strings.Split(msg, " ")
}

func (s *SoloClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscpChainID := s.ctx.Cvt.IscpChainID(&chainID)
	iscpContract := s.ctx.Cvt.IscpHname(hContract)
	iscpFunction := s.ctx.Cvt.IscpHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	if !iscpChainID.Equals(s.ctx.Chain.ChainID) {
		return reqID, errors.New("SoloClientService.PostRequest chain ID mismatch")
	}
	req := solo.NewCallParamsFromDictByHname(iscpContract, iscpFunction, params)
	s.nonce++
	req.WithNonce(s.nonce)
	iscpAllowance := s.ctx.Cvt.IscpAllowance(allowance)
	req.WithAllowance(iscpAllowance)
	req.WithGasBudget(gas.MaxGasPerCall)
	_, err = s.ctx.Chain.PostRequestOffLedger(req, keyPair)
	return reqID, err
}

func (s *SoloClientService) SubscribeEvents(msg chan []string, done chan bool) error {
	s.msg = msg
	go func() {
		<-done
	}()
	return nil
}

func (s *SoloClientService) WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	return nil
}
