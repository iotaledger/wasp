// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"time"
)

type WorkflowStatus struct {
	flagStateReceived        bool
	flagBatchProposalSent    bool
	flagConsensusBatchKnown  bool
	flagVMStarted            bool
	flagVMResultSigned       bool
	flagTransactionFinalized bool
	flagTransactionPosted    bool
	flagTransactionSeen      bool
	flagInProgress           bool

	timeBatchProposalSent    time.Time
	timeConsensusBatchKnown  time.Time
	timeVMStarted            time.Time
	timeVMResultSigned       time.Time
	timeTransactionFinalized time.Time
	timeTransactionPosted    time.Time
	timeTransactionSeen      time.Time
	timeCompleted            time.Time
}

func newWorkflowStatus(stateReceived bool) *WorkflowStatus {
	return &WorkflowStatus{
		flagStateReceived: stateReceived,
		flagInProgress:    stateReceived,
	}
}

func (wsT *WorkflowStatus) setBatchProposalSent() {
	wsT.flagBatchProposalSent = true
	wsT.timeBatchProposalSent = time.Now()
}

func (wsT *WorkflowStatus) setConsensusBatchKnown() {
	wsT.flagConsensusBatchKnown = true
	wsT.timeConsensusBatchKnown = time.Now()
}

func (wsT *WorkflowStatus) setVMStarted() {
	wsT.flagVMStarted = true
	wsT.timeVMStarted = time.Now()
}

func (wsT *WorkflowStatus) setVMResultSigned() {
	wsT.flagVMResultSigned = true
	wsT.timeVMResultSigned = time.Now()
}

func (wsT *WorkflowStatus) setTransactionFinalized() {
	wsT.flagTransactionFinalized = true
	wsT.timeTransactionFinalized = time.Now()
}

func (wsT *WorkflowStatus) setTransactionPosted() {
	wsT.flagTransactionPosted = true
	wsT.timeTransactionPosted = time.Now()
}

func (wsT *WorkflowStatus) setTransactionSeen() {
	wsT.flagTransactionSeen = true
	wsT.timeTransactionSeen = time.Now()
}

func (wsT *WorkflowStatus) setCompleted() {
	wsT.flagInProgress = false
	wsT.timeCompleted = time.Now()
}

func (wsT *WorkflowStatus) IsStateReceived() bool {
	return wsT.flagStateReceived
}

func (wsT *WorkflowStatus) IsBatchProposalSent() bool {
	return wsT.flagBatchProposalSent
}

func (wsT *WorkflowStatus) IsConsensusBatchKnown() bool {
	return wsT.flagConsensusBatchKnown
}

func (wsT *WorkflowStatus) IsVMStarted() bool {
	return wsT.flagVMStarted
}

func (wsT *WorkflowStatus) IsVMResultSigned() bool {
	return wsT.flagVMResultSigned
}

func (wsT *WorkflowStatus) IsTransactionFinalized() bool {
	return wsT.flagTransactionFinalized
}

func (wsT *WorkflowStatus) IsTransactionPosted() bool {
	return wsT.flagTransactionPosted
}

func (wsT *WorkflowStatus) IsTransactionSeen() bool {
	return wsT.flagTransactionSeen
}

func (wsT *WorkflowStatus) IsInProgress() bool {
	return wsT.flagInProgress
}

func (wsT *WorkflowStatus) GetBatchProposalSentTime() time.Time {
	return wsT.timeBatchProposalSent
}

func (wsT *WorkflowStatus) GetConsensusBatchKnownTime() time.Time {
	return wsT.timeConsensusBatchKnown
}

func (wsT *WorkflowStatus) GetVMStartedTime() time.Time {
	return wsT.timeVMStarted
}

func (wsT *WorkflowStatus) GetVMResultSignedTime() time.Time {
	return wsT.timeVMResultSigned
}

func (wsT *WorkflowStatus) GetTransactionFinalizedTime() time.Time {
	return wsT.timeTransactionFinalized
}

func (wsT *WorkflowStatus) GetTransactionPostedTime() time.Time {
	return wsT.timeTransactionPosted
}

func (wsT *WorkflowStatus) GetTransactionSeenTime() time.Time {
	return wsT.timeTransactionSeen
}

func (wsT *WorkflowStatus) GetCompletedTime() time.Time {
	return wsT.timeCompleted
}
