// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import "time"

// LogicalTime return current logical clock time on the 'solo' instance
func (env *Solo) LogicalTime() time.Time {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()
	return env.logicalTime
}

// AdvanceClockTo advances logical clock to the specific time moment in the (logical) future
func (env *Solo) AdvanceClockTo(ts time.Time) {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()
	env.advanceClockTo(ts)
}

func (env *Solo) advanceClockTo(ts time.Time) {
	if !env.logicalTime.Before(ts) {
		env.logger.Panic("can'T advance clock to the past")
	}
	env.logicalTime = ts
}

// AdvanceClockBy advances logical clock by time step
func (env *Solo) AdvanceClockBy(step time.Duration) {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()

	env.advanceClockTo(env.logicalTime.Add(step))
	env.logger.Infof("AdvanceClockBy: logical clock advanced by %v", step)
}

// ClockStep advances logical clock by time step set by SetTimeStep
func (env *Solo) ClockStep() {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()

	env.advanceClockTo(env.logicalTime.Add(env.timeStep))
	env.logger.Infof("ClockStep: logical clock advanced by %v", env.timeStep)
}

// SetTimeStep sets default time step for the 'solo' instance
func (env *Solo) SetTimeStep(step time.Duration) {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()
	env.timeStep = step
}
