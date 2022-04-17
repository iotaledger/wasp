package wasmsolo

import (
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
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

func (s *SoloClientService) CallViewByHname(chainID *iscp.ChainID, hContract, hFunction iscp.Hname, args dict.Dict) (dict.Dict, error) {
	if !chainID.Equals(s.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.CallViewByHname chain ID mismatch")
	}
	return s.ctx.Chain.CallViewByHname(hContract, hFunction, args)
}

func (s *SoloClientService) Event(msg string) {
	msg = "vmmsg " + s.ctx.ChainID().String() + " 0 " + msg
	s.msg <- strings.Split(msg, " ")
}

func (s *SoloClientService) PostRequest(chainID *iscp.ChainID, hContract, hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) (*iscp.RequestID, error) {
	if !chainID.Equals(s.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.PostRequest chain ID mismatch")
	}
	req := solo.NewCallParamsFromDictByHname(hContract, hFuncName, params)
	s.nonce++
	req.WithNonce(s.nonce)
	req.WithAllowance(allowance)
	req.WithGasBudget(gas.MaxGasPerCall)
	_, err := s.ctx.Chain.PostRequestOffLedger(req, keyPair)
	return nil, err
}

func (s *SoloClientService) SubscribeEvents(msg chan []string, done chan bool) error {
	s.msg = msg
	go func() {
		<-done
	}()
	return nil
}

func (s *SoloClientService) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) error {
	return nil
}
