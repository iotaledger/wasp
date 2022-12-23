package wasmsolo

import (
	"errors"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SoloClientService struct {
	ctx *SoloContext
	msg chan []string
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
	iscChainID := s.ctx.Cvt.IscChainID(&chainID)
	iscContract := s.ctx.Cvt.IscHname(hContract)
	iscFunction := s.ctx.Cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	if !iscChainID.Equals(s.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.CallViewByHname chain ID mismatch")
	}
	res, err := s.ctx.Chain.CallViewByHname(s.ctx.Chain.LatestBlockIndex(), iscContract, iscFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (s *SoloClientService) Event(msg string) {
	// contract tst1pqqf4qxh2w9x7rz2z4qqcvd0y8n22axsx82gqzmncvtsjqzwmhnjs438rhk | vm (contract): 89703a45: testwasmlib.test|1671671237|tst1pqqf4qxh2w9x7rz2z4qqcvd0y8n22axsx82gqzmncvtsjqzwmhnjs438rhk|Lala
	msg = "contract " + s.ctx.CurrentChainID().String() + " | vm (contract): " + isc.Hn(s.ctx.scName).String() + ": " + msg
	s.msg <- strings.Split(msg, " ")
}

func (s *SoloClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair, nonce uint64) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := s.ctx.Cvt.IscChainID(&chainID)
	iscContract := s.ctx.Cvt.IscHname(hContract)
	iscFunction := s.ctx.Cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	if !iscChainID.Equals(s.ctx.Chain.ChainID) {
		return reqID, errors.New("SoloClientService.PostRequest chain ID mismatch")
	}
	req := solo.NewCallParamsFromDictByHname(iscContract, iscFunction, params)
	req.WithNonce(nonce)
	iscAllowance := s.ctx.Cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	req.WithGasBudget(gas.MaxGasPerRequest)
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
	_ = chainID
	_ = reqID
	_ = timeout
	return nil
}
