// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/iotaledger/wasp/packages/util"
)

const TimerPullStateRetryNameConst = "PullStateRetry"
const TimerPullStateAfterStateCandidateDelayNameConst = "PullStateAfterStateCandidateDelay"
const TimerGetBlockRetryConstNameConst = "GetBlockRetry"

func NewStateManagerTimers() util.TimerParams {
	return util.NewTimerParams(
		// period of state pull retry
		util.NewTimerParam(TimerPullStateRetryNameConst, 1*time.Second),
		// how long delay state pull after state candidate received
		util.NewTimerParam(TimerPullStateAfterStateCandidateDelayNameConst, 1*time.Second),
		util.NewTimerParam(TimerGetBlockRetryConstNameConst, 3*time.Second),
	)
}
