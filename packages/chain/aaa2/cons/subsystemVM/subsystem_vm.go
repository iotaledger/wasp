// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemVM

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/bp"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

type SubsystemVM struct {
	AggregatedProposals *bp.AggregatedBatchProposals
	stateReceived       bool
	BaseAliasOutput     *isc.AliasOutputWithID
	StateBaseline       coreutil.StateBaseline
	VirtualStateAccess  state.VirtualStateAccess
	Randomness          *hashing.HashValue
	Requests            []isc.Request
	inputsReady         bool
	inputsReadyCB       func(sub *SubsystemVM) gpa.OutMessages
	outputReady         bool
	outputReadyCB       func(output *vm.VMTask) gpa.OutMessages
}

func New(
	inputsReadyCB func(sub *SubsystemVM) gpa.OutMessages,
	outputReadyCB func(output *vm.VMTask) gpa.OutMessages,
) *SubsystemVM {
	return &SubsystemVM{inputsReadyCB: inputsReadyCB, outputReadyCB: outputReadyCB}
}

func (sub *SubsystemVM) DecidedBatchProposalsReceived(aggregatedProposals *bp.AggregatedBatchProposals) gpa.OutMessages {
	if sub.AggregatedProposals != nil || aggregatedProposals == nil {
		return nil
	}
	sub.AggregatedProposals = aggregatedProposals
	return sub.tryCompleteInputs()
}

func (sub *SubsystemVM) DecidedStateReceived(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages {
	if sub.stateReceived {
		return nil
	}
	sub.stateReceived = true
	sub.BaseAliasOutput = aliasOutput
	sub.StateBaseline = stateBaseline
	sub.VirtualStateAccess = virtualStateAccess
	return sub.tryCompleteInputs()
}

func (sub *SubsystemVM) RandomnessReceived(randomness hashing.HashValue) gpa.OutMessages {
	if sub.Randomness != nil {
		return nil
	}
	sub.Randomness = &randomness
	return sub.tryCompleteInputs()
}

func (sub *SubsystemVM) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if sub.Requests != nil || requests == nil {
		return nil
	}
	sub.Requests = requests
	return sub.tryCompleteInputs()
}

func (sub *SubsystemVM) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.AggregatedProposals == nil || !sub.stateReceived || sub.Randomness == nil || sub.Requests == nil {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub)
}

func (sub *SubsystemVM) VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages {
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
func (sub *SubsystemVM) String() string {
	str := "VM"
	if sub.outputReady {
		str += "/OK"
	} else if sub.inputsReady {
		str += "/WAIT[VM to complete]"
	} else {
		wait := []string{}
		if sub.AggregatedProposals == nil {
			wait = append(wait, "AggrProposals")
		}
		if !sub.stateReceived {
			wait = append(wait, "StateFromSM")
		}
		if sub.Randomness == nil {
			wait = append(wait, "Randomness")
		}
		if sub.Requests == nil {
			wait = append(wait, "RequestsFromMP")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
