// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/chain/consensus/batchproposal"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/vm"
)

type SyncVM struct {
	c                   *consensusImpl
	aggregatedProposals *batchproposal.AggregatedBatchProposals
	chainState          state.State
	randomness          *hashing.HashValue
	requests            []isc.Request
	vmResult            *vm.VMTaskResult
	inputsReady         bool
	outputReady         bool
}

func NewSyncVM(
	c *consensusImpl,
) *SyncVM {
	return &SyncVM{c: c}
}

func (sub *SyncVM) DecidedBatchProposalsReceived(aggregatedProposals *batchproposal.AggregatedBatchProposals) gpa.OutMessages {
	if sub.aggregatedProposals != nil || aggregatedProposals == nil {
		return nil
	}
	sub.aggregatedProposals = aggregatedProposals
	msgs := gpa.NoMessages()
	msgs.AddAll(sub.tryCompleteInputs())
	msgs.AddAll(sub.tryCompleteOutputs())
	return msgs
}

func (sub *SyncVM) DecidedStateReceived(chainState state.State) gpa.OutMessages {
	if sub.chainState != nil {
		return nil
	}
	sub.chainState = chainState
	return sub.tryCompleteInputs()
}

func (sub *SyncVM) RandomnessReceived(randomness hashing.HashValue) gpa.OutMessages {
	if sub.randomness != nil {
		return nil
	}
	sub.randomness = &randomness
	return sub.tryCompleteInputs()
}

func (sub *SyncVM) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if sub.requests != nil || requests == nil {
		return nil
	}
	sub.requests = requests
	return sub.tryCompleteInputs()
}

func (sub *SyncVM) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.aggregatedProposals == nil || sub.chainState == nil || sub.randomness == nil || sub.requests == nil {
		return nil
	}
	sub.inputsReady = true
	return sub.c.uponVMInputsReceived(sub.aggregatedProposals, sub.randomness, sub.requests)
}

func (sub *SyncVM) tryCompleteOutputs() gpa.OutMessages {
	if sub.vmResult == nil || sub.aggregatedProposals == nil {
		return nil
	}
	if sub.outputReady {
		return nil
	}
	sub.outputReady = true
	return sub.c.uponVMOutputReceived(sub.vmResult, sub.aggregatedProposals)
}

func (sub *SyncVM) VMResultReceived(vmResult *vm.VMTaskResult) gpa.OutMessages {
	if sub.vmResult != nil || vmResult == nil {
		return nil
	}
	sub.vmResult = vmResult
	return sub.tryCompleteOutputs()
}

// String tries to provide useful human-readable compact status.
func (sub *SyncVM) String() string {
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
