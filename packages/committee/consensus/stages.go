package consensus

import "time"

const (
	// in the beginning
	consensusStageNoSync = iota
	// leader is just starting its activity, probably after rotation
	// no notifications has been sent yet
	consensusStageLeaderStarting
	// calculations on VM started on leader or on subordinate
	consensusStageCalculationsStarted
	// calculations on VM finished on leader or on subordinate
	consensusStageCalculationsFinished
	// finalized result has been sent to the tangle and peers. For leader only
	consensusStageResultFinalized
	// notifications has been sent to the leader. For non-leader only
	consensusStageNotificationsSent
)

type stateParams struct {
	name           string
	leaderState    bool
	nonLeaderState bool
	timeoutSet     bool
	timeout        time.Duration
	nextStages     []int
}

var stages = map[int]*stateParams{
	consensusStageNoSync: {"NoSync",
		true, true, false, 0,
		[]int{consensusStageLeaderStarting},
	},
	consensusStageLeaderStarting: {"LeaderStarting",
		true, true, false, 0,
		[]int{consensusStageCalculationsStarted, consensusStageNotificationsSent},
	},
	consensusStageCalculationsStarted: {"CalculationsStarted",
		true, true, true, 15 * time.Second,
		[]int{consensusStageCalculationsFinished},
	},
	consensusStageCalculationsFinished: {"CalculationsFinished",
		true, true, true, 5 * time.Second,
		[]int{consensusStageResultFinalized},
	},
	consensusStageResultFinalized: {"ResultFinalized",
		true, true, true, 20 * time.Second,
		[]int{},
	},
	consensusStageNotificationsSent: {"NotificationsSent",
		false, true, true, 15 * time.Second,
		[]int{consensusStageCalculationsStarted, consensusStageResultFinalized},
	},
}

func (op *operator) setConsensusStage(nextStage int) {
	currentStageParams, ok := stages[op.consensusStage]
	if !ok {
		op.log.Panicf("incorrect consensusStage 1")
	}
	nextStageParams, ok := stages[nextStage]
	if !ok {
		op.log.Panicf("incorrect consensusStage 2")
	}

	if op.iAmCurrentLeader() {
		if !nextStageParams.leaderState {
			op.log.Panicf("incorrect consensusStage 3")
		}
	} else {
		if !nextStageParams.nonLeaderState {
			op.log.Panicf("incorrect consensusStage 4")
		}
	}
	if !oneOf(nextStage, []int{consensusStageLeaderStarting, op.consensusStage}) {
		found := false
		for _, stg := range currentStageParams.nextStages {
			if stg == nextStage {
				found = true
				break
			}
		}
		if !found {
			op.log.Panicf("wrong next consensusStage: %s -> %s",
				stages[op.consensusStage].name, nextStageParams.name)
		}
	}
	saveStage := op.consensusStage
	op.consensusStage = nextStage
	op.consensusStageDeadlineSet = nextStageParams.timeoutSet
	if op.consensusStageDeadlineSet {
		op.consensusStageDeadline = time.Now().Add(nextStageParams.timeout)
	}
	op.log.Infof("consensus stage: %s -> %s, timeout in %v",
		stages[saveStage].name, nextStageParams.name, nextStageParams.timeout)
}

func (op *operator) consensusStageDeadlineExpired() bool {
	if !op.consensusStageDeadlineSet {
		return false
	}
	return time.Now().After(op.consensusStageDeadline)
}

func oneOf(elem int, set []int) bool {
	for _, e := range set {
		if e == elem {
			return true
		}
	}
	return false
}
