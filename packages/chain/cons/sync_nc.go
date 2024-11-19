package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/gpa"
)

type SyncNC interface {
	HaveState() gpa.OutMessages
	HaveRequests() gpa.OutMessages
	HaveGasInfo(gasCoins []*iotago.ObjectRef, gasPrice uint64) gpa.OutMessages
	String() string
}

type syncNCImpl struct {
	haveState    bool
	haveRequests bool
	inputCB      func() gpa.OutMessages

	gasCoins []*iotago.ObjectRef
	gasPrice uint64
	outputCB func(gasCoins []*iotago.ObjectRef, gasPrice uint64) gpa.OutMessages
}

func NewSyncNC(
	inputCB func() gpa.OutMessages,
	outputCB func(gasCoins []*iotago.ObjectRef, gasPrice uint64) gpa.OutMessages,
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
	if !sync.haveState || !sync.haveRequests || sync.inputCB == nil {
		return nil
	}
	cb := sync.inputCB
	sync.inputCB = nil
	return cb()
}

func (sync *syncNCImpl) HaveGasInfo(gasCoins []*iotago.ObjectRef, gasPrice uint64) gpa.OutMessages {
	if sync.gasCoins == nil && gasCoins != nil {
		sync.gasCoins = gasCoins
	}
	if sync.gasPrice == 0 && gasPrice != 0 {
		sync.gasPrice = gasPrice
	}
	return sync.tryCompleteOutput()
}

func (sync *syncNCImpl) tryCompleteOutput() gpa.OutMessages {
	if sync.outputCB == nil || sync.gasCoins == nil || sync.gasPrice == 0 {
		return nil
	}
	cb := sync.outputCB
	sync.outputCB = nil
	return cb(sync.gasCoins, sync.gasPrice)
}
