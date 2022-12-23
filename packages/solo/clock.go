// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"time"
)

// GlobalTime return current logical clock time on the 'solo' instance
func (env *Solo) GlobalTime() time.Time {
	return env.utxoDB.GlobalTime()
}

// AdvanceClockBy advances logical clock by time step
func (env *Solo) AdvanceClockBy(step time.Duration) {
	env.utxoDB.AdvanceClockBy(step)
	env.logger.Infof("AdvanceClockBy: logical clock advanced by %v to %s",
		step, env.utxoDB.GlobalTime().Format(timeLayout))
}
