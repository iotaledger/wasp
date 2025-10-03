// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type SyncDistributedSignature interface {
	InitialInputReceived() gpa.OutMessages
	DistributedSignatureGenerated(output gpa.Output) gpa.OutMessages
	DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) gpa.OutMessages
	MessageToSignReceived(messageToSign []byte) gpa.OutMessages
	String() string
}

type syncDistributedSignatureImpl struct {
	DecidedIndexProposals map[gpa.NodeID][]int
	MessageToSign         []byte
	initialInputsReady    bool
	initialInputsReadyCB  func() gpa.OutMessages
	indexProposalReady    bool
	indexProposalReadyCB  func(indexProposal []int) gpa.OutMessages
	signingInputsReady    bool
	signingInputsReadyCB  func(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.OutMessages
	outputReady           bool
	outputReadyCB         func(signature []byte) gpa.OutMessages
}

func NewSyncDistributedSignature(
	initialInputsReadyCB func() gpa.OutMessages,
	indexProposalReadyCB func(indexProposals []int) gpa.OutMessages,
	signingInputsReadyCB func(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.OutMessages,
	outputReadyCB func(signature []byte) gpa.OutMessages,
) SyncDistributedSignature {
	return &syncDistributedSignatureImpl{
		initialInputsReadyCB: initialInputsReadyCB,
		signingInputsReadyCB: signingInputsReadyCB,
		indexProposalReadyCB: indexProposalReadyCB,
		outputReadyCB:        outputReadyCB,
	}
}

func (sub *syncDistributedSignatureImpl) InitialInputReceived() gpa.OutMessages {
	if sub.initialInputsReady {
		return nil
	}
	sub.initialInputsReady = true
	return sub.initialInputsReadyCB()
}

func (sub *syncDistributedSignatureImpl) DistributedSignatureGenerated(output gpa.Output) gpa.OutMessages {
	if output == nil || (sub.indexProposalReady && sub.outputReady) {
		return nil
	}
	msgs := gpa.NoMessages()
	distSignOutput := output.(*distsign.Output)
	if !sub.indexProposalReady && distSignOutput.ProposedIndexes != nil {
		sub.indexProposalReady = true
		msgs.AddAll(sub.indexProposalReadyCB(distSignOutput.ProposedIndexes))
	}
	if !sub.outputReady && distSignOutput.Signature != nil {
		sub.outputReady = true
		msgs.AddAll(sub.outputReadyCB(distSignOutput.Signature))
	}
	return msgs
}

func (sub *syncDistributedSignatureImpl) DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) gpa.OutMessages {
	if sub.DecidedIndexProposals != nil || decidedIndexProposals == nil {
		return nil
	}
	sub.DecidedIndexProposals = decidedIndexProposals
	return sub.tryCompleteSigning()
}

func (sub *syncDistributedSignatureImpl) MessageToSignReceived(messageToSign []byte) gpa.OutMessages {
	if sub.MessageToSign != nil || messageToSign == nil {
		return nil
	}
	sub.MessageToSign = messageToSign
	return sub.tryCompleteSigning()
}

func (sub *syncDistributedSignatureImpl) tryCompleteSigning() gpa.OutMessages {
	if sub.signingInputsReady || sub.MessageToSign == nil || sub.DecidedIndexProposals == nil {
		return nil
	}
	sub.signingInputsReady = true
	return sub.signingInputsReadyCB(sub.DecidedIndexProposals, sub.MessageToSign)
}

// Try to provide useful human-readable compact status.
func (sub *syncDistributedSignatureImpl) String() string {
	str := "distSign"
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
		str += "/sig[WaitingFordistSign]"
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
