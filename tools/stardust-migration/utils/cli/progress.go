package cli

import (
	"strings"
	"time"
)

func NewProgressPrinter(entityPluralName string, totalCount uint32) (printProgress func(), done func()) {
	const period = time.Second
	startTime := time.Now()
	lastEstimateUpdateTime := startTime
	entityPluralNameCapitalized := strings.Title(entityPluralName)
	var avgSpeed int
	var currentSpeed int

	if totalCount == 0 {
		totalProcessed := 0
		recentlyProcessed := 0

		return func() {
				totalProcessed++
				recentlyProcessed++

				periodicAction(period, &lastEstimateUpdateTime, func() {
					avgSpeed = int(float64(totalProcessed) / time.Since(startTime).Seconds())
					currentSpeed = int(float64(recentlyProcessed) / period.Seconds())
					recentlyProcessed = 0
				})

				UpdateStatusBarf("%v processed: %v. Speed: %v %v/sec. Avg speed: %v %v/sec.",
					entityPluralNameCapitalized, totalProcessed, currentSpeed, entityPluralName, avgSpeed, entityPluralName)
			}, func() {
				UpdateStatusBarf("")
			}
	}

	var estimateRunTime time.Duration
	countLeft := totalCount
	prevCountLeft := countLeft

	return func() {
			countLeft--

			periodicAction(period, &lastEstimateUpdateTime, func() {
				totalProcessed := totalCount - countLeft
				relProgress := float64(totalProcessed) / float64(totalCount)
				estimateRunTime = time.Duration(float64(time.Since(startTime)) / relProgress)
				avgSpeed = int(float64(totalProcessed) / time.Since(startTime).Seconds())

				recentlyProcessed := prevCountLeft - countLeft
				currentSpeed = int(float64(recentlyProcessed) / period.Seconds())
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
