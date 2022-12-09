package smGPAUtils

import (
	"time"
)

type artifficialTimeProvider struct {
	now    time.Time
	timers []*timer
}

type timer struct {
	time    time.Time
	channel chan time.Time
}

var _ TimeProvider = &artifficialTimeProvider{}

func NewArtifficialTimeProvider(nowOpt ...time.Time) TimeProvider {
	var now time.Time
	if len(nowOpt) > 0 {
		now = nowOpt[0]
	} else {
		now = time.Now()
	}
	return &artifficialTimeProvider{
		now:    now,
		timers: make([]*timer, 0),
	}
}

func (atpT *artifficialTimeProvider) SetNow(now time.Time) {
	atpT.now = now
	var i int
	for i = 0; i < len(atpT.timers) && atpT.timers[i].time.Before(atpT.now); i++ {
		atpT.timers[i].channel <- atpT.now
		close(atpT.timers[i].channel)
	}
	atpT.timers = atpT.timers[i:]
}

func (atpT *artifficialTimeProvider) GetNow() time.Time {
	return atpT.now
}

func (atpT *artifficialTimeProvider) After(d time.Duration) <-chan time.Time {
	timerTime := atpT.now.Add(d)
	var i int
	for i = 0; i < len(atpT.timers) && atpT.timers[i].time.Before(timerTime); i++ {
	}
	if i == len(atpT.timers) {
		atpT.timers = append(atpT.timers, nil)
	} else {
		atpT.timers = append(atpT.timers[:i+1], atpT.timers[i:]...)
	}
	channel := make(chan time.Time, 1)
	atpT.timers[i] = &timer{
		time:    timerTime,
		channel: channel,
	}
	return channel
}

func (t *timer) String() string {
	return t.time.UTC().Format(time.RFC3339)
}
