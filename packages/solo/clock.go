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
func (env *Solo) AdvanceClockBy(step time.Duration) {
	env.utxoDB.AdvanceClockBy(step)
	t := env.utxoDB.GlobalTime()
	env.logger.Infof("AdvanceClockBy: logical clock advanced by %v to %s",
		step, t.Time.Format(timeLayout))
}
