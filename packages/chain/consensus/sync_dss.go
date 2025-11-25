// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"slices"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type SyncDistributedSignature struct {
	c                     *Consensus
	DecidedIndexProposals map[gpa.NodeID][]int
	MessageToSign         []byte
	initialInputsReady    bool
	indexProposalReady    bool
	signingInputsReady    bool
	outputReady           bool
}

func NewSyncDistributedSignature(c *Consensus) *SyncDistributedSignature {
	return &SyncDistributedSignature{c: c}
}

func (sub *SyncDistributedSignature) InitialInputReceived() []gpa.MessageOut {
	if sub.initialInputsReady {
		return nil
	}
	sub.initialInputsReady = true
	return sub.c.uponDistributedSignatureInitialInputsReady()
}

func (sub *SyncDistributedSignature) DistributedSignatureReady(output gpa.Output) []gpa.MessageOut {
	if output == nil || (sub.indexProposalReady && sub.outputReady) {
		return nil
	}
	var msgs []gpa.MessageOut
	distSignOutput := output.(*distsign.Output)
	if !sub.indexProposalReady && distSignOutput.ProposedIndexes != nil {
		sub.indexProposalReady = true
		msgs = slices.Concat(msgs, sub.c.uponDistributedSignatureIndexProposalReady(distSignOutput.ProposedIndexes))
	}
	if !sub.outputReady && distSignOutput.Signature != nil {
		sub.outputReady = true
		msgs = slices.Concat(msgs, sub.c.uponDistributedSignatureOutputReady(distSignOutput.Signature))
	}
	return msgs
}

func (sub *SyncDistributedSignature) DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) []gpa.MessageOut {
	if sub.DecidedIndexProposals != nil || decidedIndexProposals == nil {
		return nil
	}
	sub.DecidedIndexProposals = decidedIndexProposals
	return sub.tryCompleteSigning()
}

func (sub *SyncDistributedSignature) MessageToSignReceived(messageToSign []byte) []gpa.MessageOut {
	if sub.MessageToSign != nil || messageToSign == nil {
		return nil
	}
	sub.MessageToSign = messageToSign
	return sub.tryCompleteSigning()
}

func (sub *SyncDistributedSignature) tryCompleteSigning() []gpa.MessageOut {
	if sub.signingInputsReady || sub.MessageToSign == nil || sub.DecidedIndexProposals == nil {
		return nil
	}
	sub.signingInputsReady = true
	return sub.c.uponDistributedSignatureSigningInputsReceived(sub.DecidedIndexProposals, sub.MessageToSign)
}

// String tries to provide useful human-readable compact status.
func (sub *SyncDistributedSignature) String() string {
	str := "DistributedSignature"
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
		str += "/sig[WaitingForDistributedSignature]"
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
