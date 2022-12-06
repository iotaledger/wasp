// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/chain/cons/bp"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

type SyncVM interface {
	DecidedBatchProposalsReceived(aggregatedProposals *bp.AggregatedBatchProposals) gpa.OutMessages
	DecidedStateReceived(chainState state.State) gpa.OutMessages
	RandomnessReceived(randomness hashing.HashValue) gpa.OutMessages
	RequestsReceived(requests []isc.Request) gpa.OutMessages
	VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages
	String() string
}

type syncVMImpl struct {
	aggregatedProposals *bp.AggregatedBatchProposals
	chainState          state.State
	randomness          *hashing.HashValue
	requests            []isc.Request
	inputsReady         bool
	inputsReadyCB       func(aggregatedProposals *bp.AggregatedBatchProposals, chainState state.State, randomness *hashing.HashValue, requests []isc.Request) gpa.OutMessages
	outputReady         bool
	outputReadyCB       func(output *vm.VMTask) gpa.OutMessages
}

func NewSyncVM(
	inputsReadyCB func(aggregatedProposals *bp.AggregatedBatchProposals, chainState state.State, randomness *hashing.HashValue, requests []isc.Request) gpa.OutMessages,
	outputReadyCB func(output *vm.VMTask) gpa.OutMessages,
) SyncVM {
	return &syncVMImpl{inputsReadyCB: inputsReadyCB, outputReadyCB: outputReadyCB}
}

func (sub *syncVMImpl) DecidedBatchProposalsReceived(aggregatedProposals *bp.AggregatedBatchProposals) gpa.OutMessages {
	if sub.aggregatedProposals != nil || aggregatedProposals == nil {
		return nil
	}
	sub.aggregatedProposals = aggregatedProposals
	return sub.tryCompleteInputs()
}

func (sub *syncVMImpl) DecidedStateReceived(chainState state.State) gpa.OutMessages {
	if sub.chainState != nil {
		return nil
	}
	sub.chainState = chainState
	return sub.tryCompleteInputs()
}

func (sub *syncVMImpl) RandomnessReceived(randomness hashing.HashValue) gpa.OutMessages {
	if sub.randomness != nil {
		return nil
	}
	sub.randomness = &randomness
	return sub.tryCompleteInputs()
}

func (sub *syncVMImpl) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if sub.requests != nil || requests == nil {
		return nil
	}
	sub.requests = requests
	return sub.tryCompleteInputs()
}

func (sub *syncVMImpl) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.aggregatedProposals == nil || sub.chainState == nil || sub.randomness == nil || sub.requests == nil {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub.aggregatedProposals, sub.chainState, sub.randomness, sub.requests)
}

func (sub *syncVMImpl) VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages {
	if vmResult == nil {
		return nil
	}
	if sub.outputReady {
		return nil
	}
	sub.outputReady = true
	return sub.outputReadyCB(vmResult)
}

// Try to provide useful human-readable compact status.
func (sub *syncVMImpl) String() string {
	str := "VM"
	if sub.outputReady {
		str += statusStrOK
	} else if sub.inputsReady {
		str += "/WAIT[VM to complete]"
	} else {
		wait := []string{}
		if sub.aggregatedProposals == nil {
			wait = append(wait, "AggrProposals")
		}
		if sub.chainState == nil {
			wait = append(wait, "StateFromSM")
		}
		if sub.randomness == nil {
			wait = append(wait, "Randomness")
		}
		if sub.requests == nil {
			wait = append(wait, "RequestsFromMP")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
