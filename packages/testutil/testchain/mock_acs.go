package testchain

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
)

type MockedACSRunner struct {
	quorum   uint16
	sessions map[uint64]*acsSession
	log      *logger.Logger
	mutex    sync.Mutex
}

type acsSession struct {
	values    [][]byte
	callbacks []func(session uint64, values [][]byte)
	closed    bool
}

func NewMockedACSRunner(quorum uint16, log *logger.Logger) *MockedACSRunner {
	return &MockedACSRunner{
		quorum:   quorum,
		sessions: make(map[uint64]*acsSession),
		log:      log.Named("acs"),
	}
}

func (acs *MockedACSRunner) RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte)) {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	acs.log.Debugf("mockedACSRunner: started %v", sessionID)
	session, exist := acs.sessions[sessionID]
	if !exist {
		acs.log.Debugf("mockedACSRunner: creating new session %v", sessionID)
		session = &acsSession{
			values:    make([][]byte, 0),
			callbacks: make([]func(session uint64, values [][]byte), 0),
		}
		acs.sessions[sessionID] = session
	} else {
		acs.log.Debugf("mockedACSRunner: session %v is not new", sessionID)
	}
	if session.closed {
		acs.log.Debugf("mockedACSRunner: session %v is closed; returning without callbacks", sessionID)
		return
	}
	session.values = append(session.values, value)
	session.callbacks = append(session.callbacks, callback)

	acs.log.Debugf("mockedACSRunner: %v values collected out of needed %v", len(session.values), int(acs.quorum))
	if len(session.values) >= int(acs.quorum) {
		acs.log.Infof("mockedACSRunner: 'consensus' reached for sessionID %d", sessionID)
		session.closed = true
		for _, fun := range session.callbacks {
			go fun(sessionID, session.values)
		}
	}
}

func (acs *MockedACSRunner) Close() {
	// Nothing.
}
