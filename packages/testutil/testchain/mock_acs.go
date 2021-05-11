package testchain

import (
	"bytes"
	"sync"
)

type MockedAsynchronousCommonSubset struct {
	quorum         uint16
	currentSession []byte
	values         [][]byte
	onConsensus    func(values [][]byte)
	mutex          sync.Mutex
}

func NewMockedACS(quorum uint16, fun func(values [][]byte)) *MockedAsynchronousCommonSubset {
	return &MockedAsynchronousCommonSubset{
		quorum:      quorum,
		values:      make([][]byte, 0),
		onConsensus: fun,
	}
}

func (acs *MockedAsynchronousCommonSubset) ProposeValue(val []byte, session []byte) {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	if acs.currentSession == nil {
		acs.currentSession = session
	}
	if !bytes.Equal(acs.currentSession, session) {
		// ignore
		return
	}
	acs.values = append(acs.values, val)
	if len(acs.values) == int(acs.quorum) {
		v := acs.values
		acs.values = make([][]byte, 0)
		acs.currentSession = nil
		acs.onConsensus(v)
	}
}
