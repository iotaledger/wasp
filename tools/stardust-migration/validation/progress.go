package validation

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

var ProgressEnabled = true

var progresses = make(map[string]map[string][]*entityProgress)
var progressMutex = &sync.RWMutex{}
var progressPrintingThreadStarted = false

func NewProgressPrinter[Count constraints.Integer](contractName, migrationName, entityPluralName string, totalCount Count) (printProgress, done func()) {
	if !ProgressEnabled {
		return func() {}, func() {}
	}

	if !ConcurrentValidation {
		p, d := cli.NewProgressPrinter(entityPluralName, totalCount)
		return func() { p() }, d
	}

	progressMutex.Lock()
	defer progressMutex.Unlock()

	if !progressPrintingThreadStarted {
		progressPrintingThreadStarted = true
		go runMultiProgressPrinting()
	}

	contractProgresses, ok := progresses[contractName]
	if !ok {
		contractProgresses = make(map[string][]*entityProgress)
		progresses[contractName] = contractProgresses
	}

	progress := &entityProgress{
		entityPluralName: entityPluralName,
		totalCount:       int(totalCount),
		doneCount:        0,
	}

	contractProgresses[migrationName] = append(contractProgresses[migrationName], progress)

	printProgress = func() {
		progressMutex.RLock()
		defer progressMutex.RUnlock()
		progress.doneCount++
	}

	done = func() {
		progressMutex.RLock()
		defer progressMutex.RUnlock()
		progress.done = true

		for _, progress := range contractProgresses[migrationName] {
			if !progress.done {
				return
			}
		}

		printMultiProgress()
	}

	return printProgress, done
}

type entityProgress struct {
	entityPluralName string
	totalCount       int
	doneCount        int
	done             bool
}

func runMultiProgressPrinting() {
	lastPrintTime := time.Now()
	for {
		progressMutex.Lock()

		now := time.Now()
		if now.Sub(lastPrintTime) > 5*time.Second {
			printMultiProgress()
			lastPrintTime = now
		}

		updateMultiProgressStatus()

		progressMutex.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func printMultiProgress() {
	var s strings.Builder
	contractNames := lo.Keys(progresses)
	sort.Strings(contractNames)

	for _, contractName := range contractNames {
		contractProgresses := progresses[contractName]

		s.WriteString("\t")
		s.WriteString(contractName)
		s.WriteString(":\n")

		migrationNames := lo.Keys(contractProgresses)
		sort.Strings(migrationNames)
		for _, migrationName := range migrationNames {
			progresses := contractProgresses[migrationName]

			for _, progress := range progresses {
				doneStr := lo.Ternary(progress.done, "+ ", "  ")

				if progress.totalCount != 0 {
					p := int(float64(progress.doneCount) / float64(progress.totalCount) * 100)
					s.WriteString(fmt.Sprintf("\t\t%v%v: %v/%v %v (%v%%)\n", doneStr, migrationName, progress.doneCount, progress.totalCount, progress.entityPluralName, p))
				} else {
					s.WriteString(fmt.Sprintf("\t\t%v%v: %v %v\n", doneStr, migrationName, progress.doneCount, progress.entityPluralName))
				}
			}
		}
	}

	cli.DebugLogf("********** PROGRESS ************\n%v", s.String())
}

func updateMultiProgressStatus() {
	var totalProgresses int
	var doneProgresses int
	var totalWithCount int
	var doneWithCount int
	var totalWithoutCount int

	for _, contractProgresses := range progresses {
		for _, progresses := range contractProgresses {
			for _, progress := range progresses {
				totalProgresses++
				if progress.done {
					doneProgresses++
				}
				if progress.totalCount != 0 {
					totalWithCount += progress.totalCount
					if progress.done {
						doneWithCount += progress.doneCount
					}
				} else {
					totalWithoutCount += progress.doneCount
				}
			}
		}
	}

	cli.UpdateStatusBarf("Validations: %v/%v (countable: %v%%, uncountable: %v)",
		doneProgresses, totalProgresses,
		int(float64(doneWithCount)/float64(totalWithCount)*100),
		totalWithoutCount,
	)
}
