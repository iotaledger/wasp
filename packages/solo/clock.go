// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"time"
)

// GlobalTime return current logical clock time on the 'solo' instance
func (env *Solo) GlobalTime() time.Time {
	env.mockTimeMutex.RLock()
	defer env.mockTimeMutex.RUnlock()
	return env.mockTime
}

// AdvanceClockBy advances logical clock by time step
func (env *Solo) AdvanceClockBy(step time.Duration) {
	env.mockTimeMutex.Lock()
	env.mockTime = env.mockTime.Add(step)
	env.mockTimeMutex.Unlock()
	env.logger.LogInfof("AdvanceClockBy: logical clock advanced by %v to %s",
		step, env.GlobalTime().Format(timeLayout))
}
