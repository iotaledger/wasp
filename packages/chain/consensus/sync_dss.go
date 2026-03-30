// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/chain/dss"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type SyncDSS struct {
	c                     *Consensus
	DecidedIndexProposals map[gpa.NodeID][]int
	MessageToSign         []byte
	initialInputsReady    bool
	indexProposalReady    bool
	signingInputsReady    bool
	outputReady           bool
}

func NewSyncDSS(c *Consensus) *SyncDSS {
	return &SyncDSS{c: c}
}

func (sub *SyncDSS) InitialInputReceived() gpa.OutMessages {
	if sub.initialInputsReady {
		return nil
	}
	sub.initialInputsReady = true
	return sub.c.uponDSSInitialInputsReady()
}

func (sub *SyncDSS) DSSOutputReceived(output gpa.Output) gpa.OutMessages {
	if output == nil || (sub.indexProposalReady && sub.outputReady) {
		return nil
	}
	msgs := gpa.NoMessages()
	dssOutput := output.(*dss.Output)
	if !sub.indexProposalReady && dssOutput.ProposedIndexes != nil {
		sub.indexProposalReady = true
		msgs.AddAll(sub.c.uponDSSIndexProposalReady(dssOutput.ProposedIndexes))
	}
	if !sub.outputReady && dssOutput.Signature != nil {
		sub.outputReady = true
		msgs.AddAll(sub.c.uponDSSOutputReady(dssOutput.Signature))
	}
	return msgs
}

func (sub *SyncDSS) DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) gpa.OutMessages {
	if sub.DecidedIndexProposals != nil || decidedIndexProposals == nil {
		return nil
	}
	sub.DecidedIndexProposals = decidedIndexProposals
	return sub.tryCompleteSigning()
}

func (sub *SyncDSS) MessageToSignReceived(messageToSign []byte) gpa.OutMessages {
	if sub.MessageToSign != nil || messageToSign == nil {
		return nil
	}
	sub.MessageToSign = messageToSign
	return sub.tryCompleteSigning()
}

func (sub *SyncDSS) tryCompleteSigning() gpa.OutMessages {
	if sub.signingInputsReady || sub.MessageToSign == nil || sub.DecidedIndexProposals == nil {
		return nil
	}
	sub.signingInputsReady = true
	return sub.c.uponDSSSigningInputsReceived(sub.DecidedIndexProposals, sub.MessageToSign)
}

// String tries to provide useful human-readable compact status.
func (sub *SyncDSS) String() string {
	str := "DSS"
	if sub.indexProposalReady && sub.outputReady {
		return str + statusStrOK
	}
	if sub.indexProposalReady {
		str += "/idx=OK"
	} else {
		str += fmt.Sprintf("/idx[initialInputsReady=%v,indexProposalReady=%v]", sub.initialInputsReady, sub.indexProposalReady)
	}
	if sub.outputReady {
		str += "/sig=OK"
	} else if sub.signingInputsReady {
		str += "/sig[WaitingForDSS]"
	} else {
		wait := []string{}
		if sub.MessageToSign == nil {
			wait = append(wait, "MessageToSign")
		}
		if sub.DecidedIndexProposals == nil {
			wait = append(wait, "DecidedIndexProposals")
		}
		str += fmt.Sprintf("/sig=WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
