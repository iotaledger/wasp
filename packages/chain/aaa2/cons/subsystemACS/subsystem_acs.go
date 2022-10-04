// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemACS

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acs"
	"github.com/iotaledger/wasp/packages/isc"
)

// > UPON Reception of responses from Mempool, StateMgr and DSS NonceIndexes:
// >     Produce a batch proposal.
// >     Start the ACS.
type SubsystemACS struct {
	BaseAliasOutput  *isc.AliasOutputWithID
	RequestRefs      []*isc.RequestRef
	DSSIndexProposal []int
	TimeData         time.Time
	inputsReady      bool
	inputsReadyCB    func(sub *SubsystemACS) gpa.OutMessages
	outputReady      bool
	outputReadyCB    func(output map[gpa.NodeID][]byte, sub *SubsystemACS) gpa.OutMessages
	terminated       bool
	terminatedCB     func()
}

func New(
	inputsReadyCB func(sub *SubsystemACS) gpa.OutMessages,
	outputReadyCB func(output map[gpa.NodeID][]byte, sub *SubsystemACS) gpa.OutMessages,
	terminatedCB func(),
) *SubsystemACS {
	return &SubsystemACS{
		inputsReadyCB: inputsReadyCB,
		outputReadyCB: outputReadyCB,
		terminatedCB:  terminatedCB,
	}
}

func (sub *SubsystemACS) StateProposalReceived(proposedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.BaseAliasOutput != nil {
		return nil
	}
	sub.BaseAliasOutput = proposedBaseAliasOutput
	return sub.tryCompleteInput()
}

func (sub *SubsystemACS) MempoolRequestsReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if sub.RequestRefs != nil {
		return nil
	}
	sub.RequestRefs = requestRefs
	return sub.tryCompleteInput()
}

func (sub *SubsystemACS) DSSIndexProposalReceived(dssIndexProposal []int) gpa.OutMessages {
	if sub.DSSIndexProposal != nil {
		return nil
	}
	sub.DSSIndexProposal = dssIndexProposal
	return sub.tryCompleteInput()
}

func (sub *SubsystemACS) TimeDataReceived(timeData time.Time) gpa.OutMessages {
	if timeData.After(sub.TimeData) {
		sub.TimeData = timeData
		return sub.tryCompleteInput()
	}
	return nil
}

func (sub *SubsystemACS) tryCompleteInput() gpa.OutMessages {
	if sub.inputsReady || sub.BaseAliasOutput == nil || sub.RequestRefs == nil || sub.DSSIndexProposal == nil || sub.TimeData.IsZero() {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub)
}

func (sub *SubsystemACS) ACSOutputReceived(output gpa.Output) gpa.OutMessages {
	if output == nil {
		return nil
	}
	acsOutput, ok := output.(*acs.Output)
	if !ok {
		panic(xerrors.Errorf("acs returned unexpected output: %v", output))
	}
	if !sub.terminated && acsOutput.Terminated {
		sub.terminated = true
		sub.terminatedCB()
	}
	if sub.outputReady {
		return nil
	}
	sub.outputReady = true
	return sub.outputReadyCB(acsOutput.Values, sub)
}
