package smGPAUtils

import (
	"time"
)

type defaultTimeProvider struct{}

var _ TimeProvider = &defaultTimeProvider{}

func NewDefaultTimeProvider() TimeProvider     { return &defaultTimeProvider{} }
func (*defaultTimeProvider) SetNow(time.Time)  {}
func (*defaultTimeProvider) GetNow() time.Time { return time.Now() }

type artifficialTimeProvider struct {
	now time.Time
}

var _ TimeProvider = &artifficialTimeProvider{}

func NewArtifficialTimeProvider(nowOpt ...time.Time) TimeProvider {
	if len(nowOpt) == 0 {
		return &artifficialTimeProvider{now: time.Now()}
	}
	return &artifficialTimeProvider{now: nowOpt[0]}
}
func (atpT *artifficialTimeProvider) SetNow(now time.Time) { atpT.now = now }
func (atpT *artifficialTimeProvider) GetNow() time.Time    { return atpT.now }
