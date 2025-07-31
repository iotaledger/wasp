package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

type SyncNC interface {
	HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages
	HaveState() gpa.OutMessages
	HaveRequests() gpa.OutMessages
	HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages
	String() string
}

type syncNCImpl struct {
	inputAnchor         *isc.StateAnchor
	inputAnchorReceived bool
	stateReceived       bool
	requestsReceived    bool
	inputCB             func(anchor *isc.StateAnchor) gpa.OutMessages

	gasCoins []*coin.CoinWithRef
	l1params *parameters.L1Params
	outputCB func(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages
}

func NewSyncNC(
	inputCB func(anchor *isc.StateAnchor) gpa.OutMessages,
	outputCB func(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages,
) SyncNC {
	return &syncNCImpl{inputCB: inputCB, outputCB: outputCB}
}

func (sync *syncNCImpl) String() string {
	str := "NC"
	if sync.outputCB == nil {
		str += statusStrOK
	} else if sync.inputCB == nil {
		str += "/WAIT[NC to respond]"
	} else {
		wait := []string{}
		if !sync.inputAnchorReceived {
			wait = append(wait, "InputAnchor")
		}
		if !sync.stateReceived {
			wait = append(wait, "StateProposal")
		}
		if !sync.requestsReceived {
			wait = append(wait, "RequestProposals")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}

func (sync *syncNCImpl) HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages {
	if sync.inputAnchorReceived {
		return nil
	}
	sync.inputAnchor = anchor // can be nil.
	sync.inputAnchorReceived = true
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) HaveState() gpa.OutMessages {
	if sync.stateReceived {
		return nil
	}
	sync.stateReceived = true
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) HaveRequests() gpa.OutMessages {
	if sync.requestsReceived {
		return nil
	}
	sync.requestsReceived = true
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) tryCompleteInputs() gpa.OutMessages {
	if !sync.inputAnchorReceived || !sync.stateReceived || !sync.requestsReceived || sync.inputCB == nil {
		return nil
	}
	cb := sync.inputCB
	sync.inputCB = nil
	return cb(sync.inputAnchor)
}

func (sync *syncNCImpl) HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages {
	if sync.gasCoins == nil && gasCoins != nil {
		sync.gasCoins = gasCoins
	}
	if sync.l1params == nil && l1params != nil {
		sync.l1params = l1params
	}
	return sync.tryCompleteOutput()
}

func (sync *syncNCImpl) tryCompleteOutput() gpa.OutMessages {
	if sync.outputCB == nil || sync.gasCoins == nil || sync.l1params == nil {
		return nil
	}
	cb := sync.outputCB
	sync.outputCB = nil
	return cb(sync.gasCoins, sync.l1params)
}
