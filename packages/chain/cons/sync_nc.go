package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

type SyncNC interface {
	HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages
	HaveState() gpa.OutMessages
	HaveRequests() gpa.OutMessages
	HaveL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages
	String() string
}

type syncNCImpl struct {
	haveInputAnchor *isc.StateAnchor
	haveState       bool
	haveRequests    bool
	inputCB         func(anchor *isc.StateAnchor) gpa.OutMessages

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
		if sync.haveInputAnchor == nil {
			wait = append(wait, "InputAnchor")
		}
		if !sync.haveState {
			wait = append(wait, "StateProposal")
		}
		if !sync.haveRequests {
			wait = append(wait, "RequestProposals")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}

func (sync *syncNCImpl) HaveInputAnchor(anchor *isc.StateAnchor) gpa.OutMessages {
	if sync.haveInputAnchor != nil || anchor == nil {
		return nil
	}
	sync.haveInputAnchor = anchor
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) HaveState() gpa.OutMessages {
	if sync.haveState {
		return nil
	}
	sync.haveState = true
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) HaveRequests() gpa.OutMessages {
	if sync.haveRequests {
		return nil
	}
	sync.haveRequests = true
	return sync.tryCompleteInputs()
}

func (sync *syncNCImpl) tryCompleteInputs() gpa.OutMessages {
	if sync.haveInputAnchor == nil || !sync.haveState || !sync.haveRequests || sync.inputCB == nil {
		return nil
	}
	cb := sync.inputCB
	sync.inputCB = nil
	return cb(sync.haveInputAnchor)
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
