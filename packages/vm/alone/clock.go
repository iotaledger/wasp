package alone

import "time"

func (glb *Glb) LogicalTime() time.Time {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	return glb.logicalTime
}

func (glb *Glb) AdvanceClockTo(ts time.Time) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	glb.advanceClockTo(ts)
}

func (glb *Glb) advanceClockTo(ts time.Time) {
	if !glb.logicalTime.Before(ts) {
		glb.logger.Panic("can'T advance clock to the past")
	}
	glb.logicalTime = ts
}

func (glb *Glb) AdvanceClockBy(step time.Duration) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()

	glb.advanceClockTo(glb.logicalTime.Add(step))
	glb.logger.Infof("AdvanceClockBy: logical clock advanced by %v ahead", step)
}

func (glb *Glb) ClockStep() {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()

	glb.advanceClockTo(glb.logicalTime.Add(glb.timeStep))
	glb.logger.Infof("ClockStep: logical clock advanced by %v ahead", glb.timeStep)
}

func (glb *Glb) SetTimeStep(step time.Duration) {
	glb.glbMutex.Lock()
	defer glb.glbMutex.Unlock()
	glb.timeStep = step
}
