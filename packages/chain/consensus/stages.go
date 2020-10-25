package consensus

import (
	"fmt"
	"time"
)

// consensus goes through stages on the leader and on the subordinate side
// Whenever stage changes, it set the timeout. If stage doesn't change for the timeout period,
// leader is rotated.
// The file contains consensus stage tracking related settings and leader rotation timeouts

const (
	// valid for leader and subordinate
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

type stageParams struct {
	name               string
	isLeaderState      bool          // can be leader stage
	isSubordinateState bool          // can be subordinate stage
	timeoutSet         bool          // is timeout set for the stage
	timeout            time.Duration // timout value
	expectedNextStages []int         // valid next stages
}

var stages = map[int]*stageParams{
	consensusStageNoSync: {"NoSync",
		true, true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
		},
	},
	// if node knows the result is booked, it never rotates the leader in order not create conflicts
	consensusStageResultTransactionBooked: {"ResultTransactionBooked",
		true, true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
		},
	},
	// leader stages
	consensusStageLeaderStarting: {"LeaderStarting",
		true, false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderCalculationsStarted,
		},
	},
	// unlimited time for VM TODO VM timeouts/gas budget to prevent loops
	consensusStageLeaderCalculationsStarted: {"LeaderCalculationsStarted",
		true, false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderCalculationsFinished,
		},
	},
	// 30 sec for subordinates sent their signatures and finalize the result
	consensusStageLeaderCalculationsFinished: {"LeaderCalculationsFinished",
		true, false, true, 30 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageLeaderResultFinalized,
		},
	},
	// once finalized and posted result, the leader is not rotated
	consensusStageLeaderResultFinalized: {"LeaderResultFinalized",
		true, false, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageLeaderStarting,
			consensusStageSubStarting,
			consensusStageResultTransactionBooked,
		},
	},
	// subordinate stages
	// no rotation until notifications are sent
	consensusStageSubStarting: {"SubStarting",
		false, true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubNotificationsSent,
			consensusStageSubCalculationsStarted,
			consensusStageSubResultFinalized,
		},
	},
	// 30 seconds for the leader to collect notifications and send back batch of requests
	consensusStageSubNotificationsSent: {"SubNotificationsSent",
		false, true, true, 30 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubCalculationsStarted,
			consensusStageSubResultFinalized,
		},
	},
	// unlimited time for VM TODO VM timeouts/gas budget to prevent loops
	consensusStageSubCalculationsStarted: {"SubCalculationsStarted",
		false, true, false, 0,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubCalculationsFinished,
			consensusStageSubResultFinalized,
		},
	},
	// 30 sec for the leader to finalize the result
	consensusStageSubCalculationsFinished: {"SubCalculationsFinished",
		false, true, true, 30 * time.Second,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageSubResultFinalized,
			consensusStageResultTransactionBooked,
		},
	},
	// 2 minutes until 'booked' notification received
	consensusStageSubResultFinalized: {"SubResultFinalized",
		false, true, true, 2 * time.Minute,
		[]int{
			consensusStageNoSync,
			consensusStageSubStarting,
			consensusStageLeaderStarting,
			consensusStageResultTransactionBooked,
		},
	},
}

func (op *operator) mustStageParams(stage int) *stageParams {
	ret, ok := stages[stage]
	if !ok {
		op.log.Panicf("wrong stage code %d", stage)
	}
	return ret
}

// mustStageParams validates the stage with the current state
func (op *operator) mustNextStageParams(stage int) *stageParams {
	ret := op.mustStageParams(stage)
	leader, _ := op.currentLeader()
	iAmTheLeader := op.iAmCurrentLeader()
	if !ret.isLeaderState && iAmTheLeader || !ret.isSubordinateState && !iAmTheLeader {
		op.log.Panicf("stage not allowed code %d, name: '%s', leader: %d, iAmTheCurrentLeader: %v",
			stage, ret.name, leader, iAmTheLeader)
	}
	return ret
}

// setNextConsensusStage validates and sets the next stage, including timeout
func (op *operator) setNextConsensusStage(nextStage int) {
	currentStageParams := op.mustStageParams(op.consensusStage)
	nextStageParams := op.mustNextStageParams(nextStage)

	leader, _ := op.currentLeader()
	if !oneOf(nextStage, currentStageParams.expectedNextStages...) {
		op.log.Warnf("UNEXPECTED next consensusStage: %s -> %s, leader: %d, iAmTheLeader: %v",
			stages[op.consensusStage].name, nextStageParams.name, leader, op.iAmCurrentLeader())
	}
	saveStage := op.consensusStage
	op.consensusStage = nextStage
	op.consensusStageDeadline = time.Now().Add(nextStageParams.timeout)
	timeout := "timeout: not set"
	if nextStageParams.timeoutSet {
		timeout = fmt.Sprintf("timeout: %v", nextStageParams.timeout)
	}
	op.log.Debugf("consensus stage: %s -> %s, %s, leader: %d, iAmTheLeader: %v",
		stages[saveStage].name, nextStageParams.name, timeout, leader, op.iAmCurrentLeader())
}

func (op *operator) consensusStageDeadlineExpired() bool {
	stageParams := op.mustStageParams(op.consensusStage)
	if !stageParams.timeoutSet {
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
