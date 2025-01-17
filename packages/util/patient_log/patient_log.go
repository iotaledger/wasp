package patient_log

import (
	"sync"
	"time"
)

var lastLog sync.Map

func LogTimeLimited(key string, duration time.Duration, cb func()) {
	item, _ := lastLog.LoadOrStore(key, time.Now())
	t := item.(time.Time)

	if time.Since(t) > duration {
		lastLog.Store(key, time.Now())
		cb()
	}
}
