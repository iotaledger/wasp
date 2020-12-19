// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import "time"

// LogicalTime return current logical clock time on the 'solo' instance
func (glb *Solo) LogicalTime() time.Time {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	return glb.logicalTime
}

// AdvanceClockTo advances logical clock to the specific time moment in the (logical) future
func (glb *Solo) AdvanceClockTo(ts time.Time) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	glb.advanceClockTo(ts)
}

func (glb *Solo) advanceClockTo(ts time.Time) {
	if !glb.logicalTime.Before(ts) {
		glb.logger.Panic("can'T advance clock to the past")
	}
	glb.logicalTime = ts
}

// AdvanceClockBy advances logical clock by time step
func (glb *Solo) AdvanceClockBy(step time.Duration) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()

	glb.advanceClockTo(glb.logicalTime.Add(step))
	glb.logger.Infof("AdvanceClockBy: logical clock advanced by %v ahead", step)
}

// ClockStep advances logical clock by time step set by SetTimeStep
func (glb *Solo) ClockStep() {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()

	glb.advanceClockTo(glb.logicalTime.Add(glb.timeStep))
	glb.logger.Infof("ClockStep: logical clock advanced by %v ahead", glb.timeStep)
}

// SetTimeStep sets default time step for the 'solo' instance
func (glb *Solo) SetTimeStep(step time.Duration) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	glb.timeStep = step
}
