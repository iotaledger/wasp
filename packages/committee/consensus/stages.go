package consensus

import "time"

// the file contains consensus stage tracking related setting and timeouts

const (
	consensusStageNoSync = iota
	consensusStageResultTransactionBooked
	// leader stages
	consensusStageLeaderStarting
	consensusStageLeaderCalculationsStarted
	consensusStageLeaderCalculationsFinished
	consensusStageLeaderResultFinalized
	// subordinate stages
	consensusStageSubStarting
	consensusStageSubNotificationsSent
	consensusStageSubCalculationsStarted
	consensusStageSubCalculationsFinished
	consensusStageSubResultFinalized
)

type stateParams struct {
	name               string
	isLeaderState      bool
	timeoutSet         bool
	timeout            time.Duration
	expectedNextStages []int
}

var stages = map[int]*stateParams{
	consensusStageNoSync: {"NoSync",
		false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
		},
	},
	consensusStageResultTransactionBooked: {"ResultTransactionBooked",
		false, false, 1 * time.Minute,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
		},
	},
	// leader stages
	consensusStageLeaderStarting: {"LeaderStarting",
		true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderCalculationsStarted,
		},
	},
	consensusStageLeaderCalculationsStarted: {"LeaderCalculationsStarted",
		true, true, 15 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderCalculationsFinished,
		},
	},
	consensusStageLeaderCalculationsFinished: {"LeaderCalculationsFinished",
		true, true, 10 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderResultFinalized,
		},
	},
	consensusStageLeaderResultFinalized: {"LeaderResultFinalized",
		true, true, 20 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageResultTransactionBooked,
		},
	},
	// subordinate stages
	consensusStageSubStarting: {"SubStarting",
		false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubNotificationsSent,
			consensusStageSubCalculationsStarted,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubNotificationsSent: {"SubNotificationsSent",
		false, true, 15 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubCalculationsStarted,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubCalculationsStarted: {"SubCalculationsStarted",
		false, true, 15 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubCalculationsFinished,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubCalculationsFinished: {"SubCalculationsFinished",
		false, true, 5 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubResultFinalized,
			consensusStageResultTransactionBooked,
		},
	},
	consensusStageSubResultFinalized: {"SubResultFinalized",
		false, true, 20 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageResultTransactionBooked,
		},
	},
}

func (op *operator) setConsensusStage(nextStage int) {
	currentStageParams, ok := stages[op.consensusStage]
	if !ok {
		op.log.Panicf("incorrect consensusStage 1")
	}
	leader, _ := op.currentLeader()
	nextStageParams, ok := stages[nextStage]
	if !ok {
		op.log.Panicf("incorrect consensusStage 2: nextStage: %s, leader: %d, iAmTheCurrentLeader: %v",
			nextStageParams.name, leader, op.iAmCurrentLeader())
	}

	if nextStage != consensusStageNoSync && nextStageParams.isLeaderState && !op.iAmCurrentLeader() {
		op.log.Panicf("incorrect consensusStage 3: nextStage: %s, leader: %d, iAmTheCurrentLeader: %v",
			nextStageParams.name, leader, op.iAmCurrentLeader())
	}
	if !oneOf(nextStage, currentStageParams.expectedNextStages...) {
		op.log.Warnf("UNEXPECTED next consensusStage: %s -> %s, leader: %d, iAmTheLeader: %v",
			stages[op.consensusStage].name, nextStageParams.name, leader, op.iAmCurrentLeader())
	}
	saveStage := op.consensusStage
	op.consensusStage = nextStage
	op.consensusStageDeadlineSet = nextStageParams.timeoutSet
	if op.consensusStageDeadlineSet {
		op.consensusStageDeadline = time.Now().Add(nextStageParams.timeout)
	}
	op.log.Debugf("consensus stage: %s -> %s, timeout in %v, leader: %d, iAmTheLeader: %v",
		stages[saveStage].name, nextStageParams.name, nextStageParams.timeout, leader, op.iAmCurrentLeader())
}

func (op *operator) consensusStageDeadlineExpired() bool {
	if !op.consensusStageDeadlineSet {
		return false
	}
	return time.Now().After(op.consensusStageDeadline)
}

func (op *operator) setConsensusStageDeadlineExpired() {
	if !op.consensusStageDeadlineSet {
		return
	}
	op.consensusStageDeadline = time.Now()
}

func oneOf(elem int, set ...int) bool {
	for _, e := range set {
		if e == elem {
			return true
		}
	}
	return false
}
