package sm_gpa

import "time"

type (
	incFun      func()
	decFun      func()
	durationFun func(time.Duration)
)

type blockFetchersMetricsImpl struct {
	incFun      incFun
	decFun      decFun
	durationFun durationFun
}

var bfmNopDurationFun = func(time.Duration) {}

var (
	_ blockFetchersMetrics = &blockFetchersMetricsImpl{}
	_ durationFun          = bfmNopDurationFun
)

func newBlockFetchersMetrics(incFun incFun, decFun decFun, durationFun durationFun) blockFetchersMetrics {
	return &blockFetchersMetricsImpl{
		incFun:      incFun,
		decFun:      decFun,
		durationFun: durationFun,
	}
}

func (bfmiT *blockFetchersMetricsImpl) inc() {
	bfmiT.incFun()
}

func (bfmiT *blockFetchersMetricsImpl) dec() {
	bfmiT.decFun()
}

func (bfmiT *blockFetchersMetricsImpl) duration(duration time.Duration) {
	bfmiT.durationFun(duration)
}
