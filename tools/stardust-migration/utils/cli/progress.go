package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/bot"
	"golang.org/x/exp/constraints"
)

var lastNotifications map[string]time.Time = make(map[string]time.Time)
var lastGlobalStatusUpdate time.Time = time.Now()

func onlyForBlockProgress(entityPluralName string, msgType string, msg string) {
	if entityPluralName != "blocks" {
		return
	}

	// Send a fresh message every 45m so we have performance history
	if time.Now().Sub(lastGlobalStatusUpdate).Minutes() > 45 {
		lastGlobalStatusUpdate = time.Now()
		bot.Get().InvalidateStatusUpdateReference()
	}

	if time.Now().Sub(lastNotifications[msgType]).Seconds() > 30 {
		lastNotifications[msgType] = time.Now()
		bot.Get().PostStatusUpdate(fmt.Sprintf("(This will update every 30s)\nStatus Update: %s\n%v", msgType, msg))
	}
}

func NewProgressPrinter[Count constraints.Integer](entityPluralName string, totalCount Count) (printProgress func(), done func()) {
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

				updateString := fmt.Sprintf("%v processed: %v. Speed: %v %v/sec. Avg speed: %v %v/sec.",
					entityPluralNameCapitalized, totalProcessed, currentSpeed, entityPluralName, avgSpeed, entityPluralName)

				UpdateStatusBar(updateString)

				onlyForBlockProgress(entityPluralName, "processing", fmt.Sprintf("Time: %v\n"+updateString, time.Now().Format(time.DateTime)))
			}, func() {
				UpdateStatusBarf("")
			}
	}

	var estimateRunTime time.Duration
	countLeft := totalCount
	prevCountLeft := countLeft

	return func() {
			countLeft--
			totalProcessed := totalCount - countLeft

			periodicAction(period, &lastEstimateUpdateTime, func() {
				relProgress := float64(totalProcessed) / float64(totalCount)
				estimateRunTime = time.Duration(float64(time.Since(startTime)) * (1/relProgress - 1))
				avgSpeed = int(float64(totalProcessed) / time.Since(startTime).Seconds())

				recentlyProcessed := prevCountLeft - countLeft
				currentSpeed = int(float64(recentlyProcessed) / period.Seconds())
				prevCountLeft = countLeft
			})

			updateString := fmt.Sprintf("%v left: %v. Speed: %v %v/sec. Avg speed: %v %v/sec. Estimate time left: %v",
				entityPluralNameCapitalized, countLeft, currentSpeed, entityPluralName, avgSpeed, entityPluralName, estimateRunTime)

			UpdateStatusBar(updateString)

			onlyForBlockProgress(entityPluralName, "estimate", fmt.Sprintf("Block index: %v. %v", totalProcessed, updateString))
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
