package sm_snapshots

import (
	"io"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/iotaledger/hive.go/logger"
)

type progressReporter struct {
	log        *logger.Logger
	header     string
	lastReport time.Time

	expected  uint64
	total     uint64
	prevTotal uint64
}

var _ io.Writer = &progressReporter{}

const logStatusPeriodConst = 1 * time.Second

func NewProgressReporter(log *logger.Logger, header string, expected uint64) io.Writer {
	return &progressReporter{
		log:        log,
		header:     header,
		lastReport: time.Time{},
		expected:   expected,
		total:      0,
		prevTotal:  0,
	}
}

func (pr *progressReporter) Write(p []byte) (int, error) {
	pr.total += uint64(len(p))
	now := time.Now()
	timeDiff := now.Sub(pr.lastReport)
	if timeDiff >= logStatusPeriodConst {
		bps := uint64(float64(pr.total-pr.prevTotal) / timeDiff.Seconds())
		pr.log.Debugf("%s: downloaded %s of %s (%s/s)", pr.header, humanize.Bytes(pr.total), humanize.Bytes(pr.expected), humanize.Bytes(bps))
		pr.lastReport = now
		pr.prevTotal = pr.total
	}
	return len(p), nil
}
