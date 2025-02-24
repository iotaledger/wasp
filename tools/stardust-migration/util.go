package main

import (
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
)

func measureTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

func measureTimeAndPrint(descr string, f func()) {
	d := measureTime(f)
	cli.Logf("%v: %v\n", descr, d)
}

func periodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}
