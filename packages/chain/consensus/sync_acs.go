// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acs"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

// > UPON Reception of responses from Mempool, StateMgr and DistributedSignature NonceIndexes:
// >     Produce a batch proposal.
// >     Start the ACS.

type SyncACS struct {
	c *Consensus

	baseStateAnchor                   *isc.StateAnchor
	baseStateAnchorReceived           bool
	RequestRefs                       []*isc.RequestRef
	DistributedSignatureIndexProposal []int
	TimeData                          time.Time
	gasCoins                          []*coin.CoinWithRef
	l1params                          *parameters.L1Params
	l1InfoReceived                    bool

	inputsReady bool
	outputReady bool
	terminated  bool
}

func NewSyncACS(
	c *Consensus,
) *SyncACS {
	return &SyncACS{
		c: c,
	}
}

func (sub *SyncACS) StateProposalReceived(proposedBaseAnchor *isc.StateAnchor) []gpa.MessageOut {
	if sub.baseStateAnchorReceived {
		return nil
	}
	sub.baseStateAnchor = proposedBaseAnchor
	sub.baseStateAnchorReceived = true
	return sub.tryCompleteInput()
}

func (sub *SyncACS) MempoolRequestsReceived(requestRefs []*isc.RequestRef) []gpa.MessageOut {
	if sub.RequestRefs != nil {
		return nil
	}
	sub.RequestRefs = requestRefs
	return sub.tryCompleteInput()
}

func (sub *SyncACS) DistributedSignatureIndexProposalReceived(distSignIndexProposal []int) []gpa.MessageOut {
	if sub.DistributedSignatureIndexProposal != nil {
		return nil
	}
	sub.DistributedSignatureIndexProposal = distSignIndexProposal
	return sub.tryCompleteInput()
}

func (sub *SyncACS) TimeDataReceived(timeData time.Time) []gpa.MessageOut {
	if timeData.After(sub.TimeData) {
		sub.TimeData = timeData
		return sub.tryCompleteInput()
	}
	return nil
}

func (sub *SyncACS) L1InfoReceived(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) []gpa.MessageOut {
	if sub.l1InfoReceived {
		return nil
	}
	sub.gasCoins = gasCoins
	sub.l1params = l1params
	sub.l1InfoReceived = true
	return sub.tryCompleteInput()
}

func (sub *SyncACS) tryCompleteInput() []gpa.MessageOut {
	if sub.inputsReady || !sub.baseStateAnchorReceived {
		return nil
	}
	if sub.RequestRefs == nil || sub.DistributedSignatureIndexProposal == nil || sub.TimeData.IsZero() || !sub.l1InfoReceived {
		return nil
	}
	sub.inputsReady = true
	return sub.c.uponACSInputsReceived(sub.baseStateAnchor, sub.RequestRefs, sub.DistributedSignatureIndexProposal, sub.TimeData, sub.gasCoins, sub.l1params)
}

func (sub *SyncACS) ACSOutputReceived(output gpa.Output) []gpa.MessageOut {
	if output == nil {
		return nil
	}
	acsOutput, ok := output.(*acs.Output)
	if !ok {
		panic(fmt.Errorf("acs returned unexpected output: %v", output))
	}
	if !sub.terminated && acsOutput.Terminated {
		sub.terminated = true
		sub.c.uponACSTerminated()
	}
	if sub.outputReady {
		return nil
	}
	sub.outputReady = true
	return sub.c.uponACSOutputReceived(acsOutput.Values)
}

// String tries to provide useful human-readable compact status.
func (sub *SyncACS) String() string {
	str := "ACS"
	if sub.outputReady {
		str += statusStrOK
	} else if sub.inputsReady {
		str += "/WAIT[ACS to complete]"
	} else {
		wait := []string{}
		if !sub.baseStateAnchorReceived {
			wait = append(wait, "BaseStateAnchor")
		}
		if sub.RequestRefs == nil {
			wait = append(wait, "RequestRefs")
		}
		if sub.DistributedSignatureIndexProposal == nil {
			wait = append(wait, "DistributedSignatureIndexProposal")
		}
		if sub.TimeData.IsZero() {
			wait = append(wait, "TimeData")
		}
		if !sub.l1InfoReceived {
			wait = append(wait, "L1Info")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
