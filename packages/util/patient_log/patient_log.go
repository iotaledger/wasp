package patient_log

import "time"

var lastLog = map[string]time.Time{}

func LogTimeLimited(key string, duration time.Duration, cb func()) {
	t, ok := lastLog[key]
	if !ok {
		lastLog[key] = time.Now()
		t = lastLog[key]
	}

	if time.Since(t) > duration {
		cb()
		lastLog[key] = time.Now()
	}
}
