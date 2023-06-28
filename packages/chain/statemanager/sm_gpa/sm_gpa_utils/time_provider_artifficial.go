package sm_gpa_utils

import (
	"sync"
	"time"
)

type artifficialTimeProvider struct {
	now    time.Time
	timers []*timer
	mutex  sync.Mutex
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
		mutex:  sync.Mutex{},
	}
}

func (atpT *artifficialTimeProvider) SetNow(now time.Time) {
	atpT.mutex.Lock()
	defer atpT.mutex.Unlock()

	atpT.now = now
	var i int
	for i = 0; i < len(atpT.timers) && atpT.timers[i].time.Before(atpT.now); i++ {
		atpT.timers[i].channel <- atpT.now
		close(atpT.timers[i].channel)
	}
	atpT.timers = atpT.timers[i:]
}

func (atpT *artifficialTimeProvider) GetNow() time.Time {
	atpT.mutex.Lock()
	defer atpT.mutex.Unlock()

	return atpT.now
}

func (atpT *artifficialTimeProvider) After(d time.Duration) <-chan time.Time {
	channel := make(chan time.Time, 1)
	if d == 0 {
		channel <- atpT.now
		close(channel)
	} else {
		atpT.mutex.Lock()
		defer atpT.mutex.Unlock()

		timerTime := atpT.now.Add(d)

		var count int
		for i := 0; i < len(atpT.timers) && atpT.timers[i].time.Before(timerTime); i++ {
			count++
		}

		if count == len(atpT.timers) {
			atpT.timers = append(atpT.timers, nil)
		} else {
			atpT.timers = append(atpT.timers[:count+1], atpT.timers[count:]...)
		}
		atpT.timers[count] = &timer{
			time:    timerTime,
			channel: channel,
		}
	}
	return channel
}

func (t *timer) String() string {
	return t.time.UTC().Format(time.RFC3339)
}
