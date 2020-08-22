package consensus

import "time"

const (
	consensusStageNoSync = iota
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
	name          string
	isLeaderState bool
	timeoutSet    bool
	timeout       time.Duration
	nextStages    []int
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
	// leader stages
	consensusStageLeaderStarting: {"LeaderStarting",
		true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageLeaderCalculationsStarted,
		},
	},
	consensusStageLeaderCalculationsStarted: {"LeaderCalculationsStarted",
		true, true, 15 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageLeaderCalculationsFinished,
		},
	},
	consensusStageLeaderCalculationsFinished: {"LeaderCalculationsFinished",
		true, true, 5 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageLeaderResultFinalized,
		},
	},
	consensusStageLeaderResultFinalized: {"LeaderResultFinalized",
		true, true, 20 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
		},
	},
	// subordinate stages
	consensusStageSubStarting: {"SubStarting",
		false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
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
			consensusStageSubCalculationsStarted,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubCalculationsStarted: {"SubCalculationsStarted",
		false, true, 15 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageSubCalculationsFinished,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubCalculationsFinished: {"SubCalculationsFinished",
		false, true, 5 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageSubResultFinalized,
		},
	},
	consensusStageSubResultFinalized: {"SubResultFinalized",
		false, true, 20 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
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
	if !oneOf(nextStage, currentStageParams.nextStages...) {
		op.log.Panicf("wrong next consensusStage: %s -> %s, leader: %d, iAmTheLeader: %v",
			stages[op.consensusStage].name, nextStageParams.name, leader, op.iAmCurrentLeader())
	}
	saveStage := op.consensusStage
	op.consensusStage = nextStage
	op.consensusStageDeadlineSet = nextStageParams.timeoutSet
	if op.consensusStageDeadlineSet {
		op.consensusStageDeadline = time.Now().Add(nextStageParams.timeout)
	}
	op.log.Infof("consensus stage: %s -> %s, timeout in %v, leader: %d, iAmTheLeader: %v",
		stages[saveStage].name, nextStageParams.name, nextStageParams.timeout, leader, op.iAmCurrentLeader())
}

func (op *operator) consensusStageDeadlineExpired() bool {
	if !op.consensusStageDeadlineSet {
		return false
	}
	return time.Now().After(op.consensusStageDeadline)
}

func oneOf(elem int, set ...int) bool {
	for _, e := range set {
		if e == elem {
			return true
		}
	}
	return false
}
