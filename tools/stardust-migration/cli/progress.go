package cli

import (
	"strings"
	"time"
)

func NewProgressPrinter(entityPluralName string, totalCount uint32) (printProgress func(), done func()) {
	countLeft := totalCount
	entityPluralNameCapitalized := strings.Title(entityPluralName)

	var estimateRunTime time.Duration
	var avgSpeed int
	var currentSpeed int
	prevCountLeft := countLeft
	startTime := time.Now()
	lastEstimateUpdateTime := time.Now()

	return func() {
			countLeft--

			const period = time.Second
			periodicAction(period, &lastEstimateUpdateTime, func() {
				totalProcessed := totalCount - countLeft
				relProgress := float64(totalProcessed) / float64(totalCount)
				estimateRunTime = time.Duration(float64(time.Since(startTime)) / relProgress)
				avgSpeed = int(float64(totalProcessed) / time.Since(startTime).Seconds())

				recentProcessed := prevCountLeft - countLeft
				currentSpeed = int(float64(recentProcessed) / period.Seconds())
				prevCountLeft = countLeft
			})

			UpdateStatusBarf("%v left: %v. Speed: %v %v/sec. Avg speed: %v %v/sec. Estimate time left: %v",
				entityPluralNameCapitalized, countLeft, currentSpeed, entityPluralName, avgSpeed, entityPluralName, estimateRunTime)
		}, func() {
			UpdateStatusBarf("")
		}
}

func periodicAction(period time.Duration, lastActionTime *time.Time, action func()) {
	if time.Since(*lastActionTime) >= period {
		action()
		*lastActionTime = time.Now()
	}
}
