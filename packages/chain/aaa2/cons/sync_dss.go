// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/gpa"
)

type SyncDSS interface {
	InitialInputReceived() gpa.OutMessages
	DSSOutputReceived(output gpa.Output) gpa.OutMessages
	DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) gpa.OutMessages
	MessageToSignReceived(messageToSign []byte) gpa.OutMessages
	String() string
}

type syncDSSImpl struct {
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

func NewSyncDSS(
	initialInputsReadyCB func() gpa.OutMessages,
	indexProposalReadyCB func(indexProposals []int) gpa.OutMessages,
	signingInputsReadyCB func(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.OutMessages,
	outputReadyCB func(signature []byte) gpa.OutMessages,
) SyncDSS {
	return &syncDSSImpl{
		initialInputsReadyCB: initialInputsReadyCB,
		signingInputsReadyCB: signingInputsReadyCB,
		indexProposalReadyCB: indexProposalReadyCB,
		outputReadyCB:        outputReadyCB,
	}
}

func (sub *syncDSSImpl) InitialInputReceived() gpa.OutMessages {
	if sub.initialInputsReady {
		return nil
	}
	sub.initialInputsReady = true
	return sub.initialInputsReadyCB()
}

func (sub *syncDSSImpl) DSSOutputReceived(output gpa.Output) gpa.OutMessages {
	if output == nil || (sub.indexProposalReady && sub.outputReady) {
		return nil
	}
	msgs := gpa.NoMessages()
	dssOutput := output.(*dss.Output)
	if !sub.indexProposalReady && dssOutput.ProposedIndexes != nil {
		sub.indexProposalReady = true
		msgs.AddAll(sub.indexProposalReadyCB(dssOutput.ProposedIndexes))
	}
	if !sub.outputReady && dssOutput.Signature != nil {
		sub.outputReady = true
		msgs.AddAll(sub.outputReadyCB(dssOutput.Signature))
	}
	return msgs
}

func (sub *syncDSSImpl) DecidedIndexProposalsReceived(decidedIndexProposals map[gpa.NodeID][]int) gpa.OutMessages {
	if sub.DecidedIndexProposals != nil || decidedIndexProposals == nil {
		return nil
	}
	sub.DecidedIndexProposals = decidedIndexProposals
	return sub.tryCompleteSigning()
}

func (sub *syncDSSImpl) MessageToSignReceived(messageToSign []byte) gpa.OutMessages {
	if sub.MessageToSign != nil || messageToSign == nil {
		return nil
	}
	sub.MessageToSign = messageToSign
	return sub.tryCompleteSigning()
}

func (sub *syncDSSImpl) tryCompleteSigning() gpa.OutMessages {
	if sub.signingInputsReady || sub.MessageToSign == nil || sub.DecidedIndexProposals == nil {
		return nil
	}
	sub.signingInputsReady = true
	return sub.signingInputsReadyCB(sub.DecidedIndexProposals, sub.MessageToSign)
}

// Try to provide useful human-readable compact status.
func (sub *syncDSSImpl) String() string {
	str := "DSS"
	if sub.indexProposalReady && sub.outputReady {
		return str + "/OK"
	}
	if sub.indexProposalReady {
		str += "/idx=OK"
	} else {
		str += fmt.Sprintf("/idx[initialInputsReady=%v]", sub.initialInputsReady)
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
