package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

type SyncNodeconn struct {
	c *Consensus

	inputProcessed      bool
	inputAnchor         *isc.StateAnchor
	inputAnchorReceived bool
	stateReceived       bool
	requestsReceived    bool

	outputProcessed bool
	gasCoins        []*coin.CoinWithRef
	l1params        *parameters.L1Params
}

func NewSyncNodeconn(c *Consensus) *SyncNodeconn {
	return &SyncNodeconn{c: c}
}

func (s *SyncNodeconn) String() string {
	str := "NC"
	if s.outputProcessed {
		str += statusStrOK
	} else if s.inputProcessed {
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

func (s *SyncNodeconn) HaveInputAnchor(anchor *isc.StateAnchor) []gpa.PayloadOut {
	if s.inputAnchorReceived {
		return nil
	}
	s.inputAnchor = anchor // can be nil.
	s.inputAnchorReceived = true
	return s.tryCompleteInputs()
}

func (s *SyncNodeconn) HaveState() []gpa.PayloadOut {
	if s.stateReceived {
		return nil
	}
	s.stateReceived = true
	return s.tryCompleteInputs()
}

func (s *SyncNodeconn) HaveRequests() []gpa.PayloadOut {
	if s.requestsReceived {
		return nil
	}
	s.requestsReceived = true
	return s.tryCompleteInputs()
}

func (s *SyncNodeconn) tryCompleteInputs() []gpa.PayloadOut {
	if !s.inputAnchorReceived || !s.stateReceived || !s.requestsReceived || s.inputProcessed {
		return nil
	}
	s.inputProcessed = true
	return s.c.uponNodeconnInputsReady(s.inputAnchor)
}

func (s *SyncNodeconn) HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) []gpa.PayloadOut {
	if s.gasCoins == nil && gasCoins != nil {
		s.gasCoins = gasCoins
	}
	if s.l1params == nil && l1params != nil {
		s.l1params = l1params
	}
	return s.tryCompleteOutput()
}

func (s *SyncNodeconn) tryCompleteOutput() []gpa.PayloadOut {
	if s.outputProcessed || s.gasCoins == nil || s.l1params == nil {
		return nil
	}
	s.outputProcessed = true
	return s.c.uponNodeconnOutputReady(s.gasCoins, s.l1params)
}
