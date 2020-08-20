package consensus

import "time"

const (
	// in the beginning
	stateInit = iota
	// SC state just changed
	stateSCStateChanged
	// leader just changed
	stateLeaderChanged
	//// leader states
	// requests to start calculations hash been sent to peers
	stateStartCalculationsOrderSent
	// finalized result has been sent to the tangle and peers
	stateResultFinalized
	//// non-leader states
	// notifications has been sent to the leader
	stateNotificationsSent
	// calculation result has been sent to the leader
	stateResultSent
	// calculation evidence received from the leader
	stateCalcEvidenceFromTheLeader
	// calculation evidence received from the tangle not confirmed yet)
	stateCalcEvidenceFromTheTangle
)

type stateParams struct {
	leaderState    bool
	nonLeaderState bool
	nextStates     map[int]byte
}

var states = map[int]*stateParams{
	stateInit: {true, true,
		map[int]byte{stateInit: 0, stateSCStateChanged: 0},
	},
	stateSCStateChanged: {true, true,
		map[int]byte{stateLeaderChanged: 0, stateSCStateChanged: 0},
	},
	stateLeaderChanged: {true, true,
		map[int]byte{stateStartCalculationsOrderSent: 0, stateNotificationsSent: 0},
	},
	stateStartCalculationsOrderSent: {true, false,
		map[int]byte{stateResultFinalized: 0},
	},
	stateResultFinalized: {true, false,
		map[int]byte{stateSCStateChanged: 0, stateLeaderChanged: 0},
	},
	stateNotificationsSent: {false, true,
		map[int]byte{stateResultSent: 0, stateLeaderChanged: 0},
	},
	stateResultSent: {false, true,
		map[int]byte{stateCalcEvidenceFromTheLeader: 0, stateLeaderChanged: 0},
	},
	stateCalcEvidenceFromTheLeader: {false, true,
		map[int]byte{stateLeaderChanged: 0, stateSCStateChanged: 0},
	},
	stateCalcEvidenceFromTheTangle: {false, true,
		map[int]byte{stateLeaderChanged: 0, stateSCStateChanged: 0},
	},
}

func (op *operator) setNextState(nextState int) {
	nextStateParams, ok := states[op.state]
	if !ok {
		op.log.Panicf("incorrect state 1")
	}

	if op.iAmCurrentLeader() {
		if !nextStateParams.leaderState {
			op.log.Panicf("incorrect state 2")
		}
	} else {
		if !nextStateParams.nonLeaderState {
			op.log.Panicf("incorrect state 3")
		}
	}
	if _, ok := nextStateParams.nextStates[nextState]; !ok {
		op.log.Panicf("wrong next state: %d -> %d", op.state, nextState)
	}
	op.state = nextState
	op.whenSetState = time.Now()
}
