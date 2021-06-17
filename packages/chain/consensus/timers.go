// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"time"

	"github.com/iotaledger/wasp/packages/util"
)

const TimerVMRunRetryToWaitForReadyRequestsNameConst = "VMRunRetryToWaitForReadyRequests"
const TimerBroadcastSignedResultRetryNameConst = "BroadcastSignedResultRetry"
const TimerPostTxSequenceStepNameConst = "PostTxSequenceStep"
const TimerPullInclusionStateRetryNameConst = "PullInclusionStateRetry"
const TimerProposeBatchRetryNameConst = "ProposeBatchRetry"

func NewConsensusTimers() util.TimerParams {
	return util.NewTimerParams(
		util.NewTimerParam(TimerVMRunRetryToWaitForReadyRequestsNameConst, 500*time.Millisecond),
		util.NewTimerParam(TimerBroadcastSignedResultRetryNameConst, 1*time.Second),
		util.NewTimerParam(TimerPostTxSequenceStepNameConst, 1*time.Second),
		util.NewTimerParam(TimerPullInclusionStateRetryNameConst, 1*time.Second),
		util.NewTimerParam(TimerProposeBatchRetryNameConst, 500*time.Millisecond),
	)
}
