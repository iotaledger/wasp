package testchain

import (
	"bytes"
	"sync"

	"github.com/iotaledger/hive.go/logger"
)

type mockedACSRunner struct {
	quorum         uint16
	currentSession []byte
	values         [][]byte
	callbacks      []func(session []byte, values [][]byte)
	log            *logger.Logger
	mutex          sync.Mutex
}

func NewMockedACSRunner(quorum uint16, log *logger.Logger) *mockedACSRunner {
	return &mockedACSRunner{quorum: quorum, log: log.Named("acs")}
}

func (acs *mockedACSRunner) RunACSConsensus(value []byte, sessionID []byte, callback func(sessionID []byte, acs [][]byte)) {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	if acs.currentSession == nil {
		acs.currentSession = sessionID
		acs.values = make([][]byte, 0, acs.quorum)
		acs.callbacks = make([]func(session []byte, values [][]byte), 0, acs.quorum)
	}
	if !bytes.Equal(acs.currentSession, sessionID) {
		// ignore
		return
	}
	acs.values = append(acs.values, value)
	acs.callbacks = append(acs.callbacks, callback)

	if len(acs.values) == int(acs.quorum) {
		acs.log.Infof("mockedACSRunner: 'consensus' reached")
		for _, fun := range acs.callbacks {
			go fun(acs.currentSession, acs.values)
		}
		acs.values = make([][]byte, 0, acs.quorum)
		acs.callbacks = make([]func(session []byte, values [][]byte), 0, acs.quorum)
	}
}
