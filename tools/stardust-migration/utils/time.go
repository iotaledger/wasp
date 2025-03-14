package utils

import (
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

func MeasureTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

func MeasureTimeAndPrint(descr string, f func()) {
	d := MeasureTime(f)
	cli.Logf("%v: %v\n", descr, d)
}

func PeriodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}
