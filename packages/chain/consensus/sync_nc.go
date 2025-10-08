package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

type SyncNodeconn interface {
	HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages
	HaveState() gpa.OutMessages
	HaveRequests() gpa.OutMessages
	HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages
	String() string
}

type syncNodeconnImpl struct {
	inputAnchor         *isc.StateAnchor
	inputAnchorReceived bool
	stateReceived       bool
	requestsReceived    bool
	inputCB             func(anchor *isc.StateAnchor) gpa.OutMessages

	gasCoins []*coin.CoinWithRef
	l1params *parameters.L1Params
	outputCB func(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages
}

func NewSyncNodeconn(
	inputCB func(anchor *isc.StateAnchor) gpa.OutMessages,
	outputCB func(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages,
) SyncNodeconn {
	return &syncNodeconnImpl{inputCB: inputCB, outputCB: outputCB}
}

func (s *syncNodeconnImpl) String() string {
	str := "NC"
	if s.outputCB == nil {
		str += statusStrOK
	} else if s.inputCB == nil {
		str += "/WAIT[NC to respond]"
	} else {
		wait := []string{}
		if !s.inputAnchorReceived {
			wait = append(wait, "InputAnchor")
		}
		if !s.stateReceived {
			wait = append(wait, "StateProposal")
		}
		if !s.requestsReceived {
			wait = append(wait, "RequestProposals")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}

func (s *syncNodeconnImpl) HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages {
	if s.inputAnchorReceived {
		return nil
	}
	s.inputAnchor = anchor // can be nil.
	s.inputAnchorReceived = true
	return s.tryCompleteInputs()
}

func (s *syncNodeconnImpl) HaveState() gpa.OutMessages {
	if s.stateReceived {
		return nil
	}
	s.stateReceived = true
	return s.tryCompleteInputs()
}

func (s *syncNodeconnImpl) HaveRequests() gpa.OutMessages {
	if s.requestsReceived {
		return nil
	}
	s.requestsReceived = true
	return s.tryCompleteInputs()
}

func (s *syncNodeconnImpl) tryCompleteInputs() gpa.OutMessages {
	if !s.inputAnchorReceived || !s.stateReceived || !s.requestsReceived || s.inputCB == nil {
		return nil
	}
	cb := s.inputCB
	s.inputCB = nil
	return cb(s.inputAnchor)
}

func (s *syncNodeconnImpl) HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages {
	if s.gasCoins == nil && gasCoins != nil {
		s.gasCoins = gasCoins
	}
	if s.l1params == nil && l1params != nil {
		s.l1params = l1params
	}
	return s.tryCompleteOutput()
}

func (s *syncNodeconnImpl) tryCompleteOutput() gpa.OutMessages {
	if s.outputCB == nil || s.gasCoins == nil || s.l1params == nil {
		return nil
	}
	cb := s.outputCB
	s.outputCB = nil
	return cb(s.gasCoins, s.l1params)
}
