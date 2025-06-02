// Package testchain provides utilities for testing chain operations and behaviors
package testchain

import (
	"sync"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/kv/codec"
)

type MockedACSRunner struct {
	quorum   uint16
	sessions map[uint64]*acsSession
	log      log.Logger
	mutex    sync.Mutex
}

type acsSession struct {
	validators map[uint16]bool
	values     [][]byte
	callbacks  []func(session uint64, values [][]byte)
	closed     bool
}

func NewMockedACSRunner(quorum uint16, log log.Logger) *MockedACSRunner {
	return &MockedACSRunner{
		quorum:   quorum,
		sessions: make(map[uint64]*acsSession),
		log:      log.NewChildLogger("acs"),
	}
}

func (acs *MockedACSRunner) RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte)) {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	acs.log.LogDebugf("mockedACSRunner: started %v", sessionID)
	session, exist := acs.sessions[sessionID]
	if !exist {
		acs.log.LogDebugf("mockedACSRunner: creating new session %v", sessionID)
		session = &acsSession{
			validators: make(map[uint16]bool),
			values:     make([][]byte, 0),
			callbacks:  make([]func(session uint64, values [][]byte), 0),
		}
		acs.sessions[sessionID] = session
	} else {
		acs.log.LogDebugf("mockedACSRunner: session %v is not new", sessionID)
	}
	if session.closed {
		acs.log.LogDebugf("mockedACSRunner: session %v is closed; returning without callbacks", sessionID)
		return
	}

	validator, err := codec.Decode[uint16](value)
	if err != nil {
		acs.log.LogErrorf("mockedACSRunner: cannot retrieve validator from batch proposal: %v", err)
		return
	}
	if session.validators[validator] {
		acs.log.LogDebugf("mockedACSRunner: batch proposal from %v is already present", err)
		return
	}

	session.values = append(session.values, value)
	session.callbacks = append(session.callbacks, callback)
	session.validators[validator] = true

	acs.log.LogDebugf("mockedACSRunner: %v values collected out of needed %v", len(session.values), int(acs.quorum))
	if len(session.values) >= int(acs.quorum) {
		acs.log.LogInfof("mockedACSRunner: 'consensus' reached for sessionID %d", sessionID)
		session.closed = true
		for _, fun := range session.callbacks {
			go fun(sessionID, session.values)
		}
	}
}

func (acs *MockedACSRunner) Close() {
	// Nothing.
}
