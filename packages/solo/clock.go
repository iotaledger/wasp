// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

// GlobalTime return current logical clock time on the 'solo' instance
func (env *Solo) GlobalTime() iscp.TimeData {
	return env.utxoDB.GlobalTime()
}

// AdvanceClockBy advances logical clock by time step
func (env *Solo) AdvanceClockBy(step time.Duration, milestones uint32) {
	env.utxoDB.AdvanceClockBy(step, milestones)
	t := env.utxoDB.GlobalTime()
	env.logger.Infof("AdvanceClockBy: logical clock advanced by %v to %s by %d milestone indices",
		step, t.Time.Format(timeLayout), t.MilestoneIndex)
}
