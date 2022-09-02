// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
)

type workflowStatus struct {
	flagStateReceived         bool
	flagIndexProposalReceived bool
	flagBatchProposalSent     bool
	flagConsensusBatchKnown   bool
	flagVMStarted             bool
	flagDssSigningStarted     bool
	flagVMResultSigned        bool
	flagTransactionFinalized  bool
	flagTransactionPosted     bool
	flagTransactionSeen       bool
	flagInProgress            bool

	timeIndexProposalReceived time.Time
	timeBatchProposalSent     time.Time
	timeConsensusBatchKnown   time.Time
	timeVMStarted             time.Time
	timeDssSigningStarted     time.Time
	timeVMResultSigned        time.Time
	timeTransactionFinalized  time.Time
	timeTransactionPosted     time.Time
	timeTransactionSeen       time.Time
	timeCompleted             time.Time

	stateIndex uint32
}

var _ chain.ConsensusWorkflowStatus = &workflowStatus{}

func newWorkflowStatus(stateReceived bool, stateIndex ...uint32) *workflowStatus {
	var i uint32
	if len(stateIndex) > 0 {
		i = stateIndex[0]
	}
	return &workflowStatus{
		flagStateReceived: stateReceived,
		flagInProgress:    stateReceived,
		stateIndex:        i,
	}
}

func (wsT *workflowStatus) setIndexProposalReceived() {
	wsT.flagIndexProposalReceived = true
	wsT.timeIndexProposalReceived = time.Now()
}

func (wsT *workflowStatus) setBatchProposalSent() {
	wsT.flagBatchProposalSent = true
	wsT.timeBatchProposalSent = time.Now()
}

func (wsT *workflowStatus) setConsensusBatchKnown() {
	wsT.flagConsensusBatchKnown = true
	wsT.timeConsensusBatchKnown = time.Now()
}

func (wsT *workflowStatus) setVMStarted() {
	wsT.flagVMStarted = true
	wsT.timeVMStarted = time.Now()
}

func (wsT *workflowStatus) setDssSigningStarted() {
	wsT.flagDssSigningStarted = true
	wsT.timeDssSigningStarted = time.Now()
}

func (wsT *workflowStatus) setVMResultSigned() {
	wsT.flagVMResultSigned = true
	wsT.timeVMResultSigned = time.Now()
}

func (wsT *workflowStatus) setTransactionFinalized() {
	wsT.flagTransactionFinalized = true
	wsT.timeTransactionFinalized = time.Now()
}

func (wsT *workflowStatus) setTransactionPosted() {
	wsT.flagTransactionPosted = true
	wsT.timeTransactionPosted = time.Now()
}

func (wsT *workflowStatus) setTransactionSeen() {
	wsT.flagTransactionSeen = true
	wsT.timeTransactionSeen = time.Now()
}

func (wsT *workflowStatus) setCompleted() {
	wsT.flagInProgress = false
	wsT.timeCompleted = time.Now()
}

func (wsT *workflowStatus) setCurrentStateIndex(i uint32) {
	wsT.stateIndex = i
}

func (wsT *workflowStatus) IsStateReceived() bool {
	return wsT.flagStateReceived
}

func (wsT *workflowStatus) IsIndexProposalReceived() bool {
	return wsT.flagIndexProposalReceived
}

func (wsT *workflowStatus) IsBatchProposalSent() bool {
	return wsT.flagBatchProposalSent
}

func (wsT *workflowStatus) IsConsensusBatchKnown() bool {
	return wsT.flagConsensusBatchKnown
}

func (wsT *workflowStatus) IsVMStarted() bool {
	return wsT.flagVMStarted
}

func (wsT *workflowStatus) IsDssSigningStarted() bool {
	return wsT.flagDssSigningStarted
}

func (wsT *workflowStatus) IsVMResultSigned() bool {
	return wsT.flagVMResultSigned
}

func (wsT *workflowStatus) IsTransactionFinalized() bool {
	return wsT.flagTransactionFinalized
}

func (wsT *workflowStatus) IsTransactionPosted() bool {
	return wsT.flagTransactionPosted
}

func (wsT *workflowStatus) IsTransactionSeen() bool {
	return wsT.flagTransactionSeen
}

func (wsT *workflowStatus) IsInProgress() bool {
	return wsT.flagInProgress
}

func (wsT *workflowStatus) GetIndexProposalReceived() time.Time {
	return wsT.timeIndexProposalReceived
}

func (wsT *workflowStatus) GetBatchProposalSentTime() time.Time {
	return wsT.timeBatchProposalSent
}

func (wsT *workflowStatus) GetConsensusBatchKnownTime() time.Time {
	return wsT.timeConsensusBatchKnown
}

func (wsT *workflowStatus) GetVMStartedTime() time.Time {
	return wsT.timeVMStarted
}

func (wsT *workflowStatus) GetDssSigningStartedTime() time.Time {
	return wsT.timeDssSigningStarted
}

func (wsT *workflowStatus) GetVMResultSignedTime() time.Time {
	return wsT.timeVMResultSigned
}

func (wsT *workflowStatus) GetTransactionFinalizedTime() time.Time {
	return wsT.timeTransactionFinalized
}

func (wsT *workflowStatus) GetTransactionPostedTime() time.Time {
	return wsT.timeTransactionPosted
}

func (wsT *workflowStatus) GetTransactionSeenTime() time.Time {
	return wsT.timeTransactionSeen
}

func (wsT *workflowStatus) GetCompletedTime() time.Time {
	return wsT.timeCompleted
}

func (wsT *workflowStatus) GetCurrentStateIndex() uint32 {
	return wsT.stateIndex
}
