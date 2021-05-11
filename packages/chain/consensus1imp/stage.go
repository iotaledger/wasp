package consensus1imp

import "fmt"

type consensusStage byte

const (
	_ = consensusStage(iota)
	stageStateReceived
	stageConsensus
	stageConsensusCompleted
	stageVM
	stageWaitForSignatures
	stageTransactionFinalized
)

func (s consensusStage) String() string {
	switch s {
	case stageStateReceived:
		return "stateReceived"
	case stageConsensus:
		return "consensusInProgress"
	case stageConsensusCompleted:
		return "consensusCompleted"
	case stageVM:
		return "VMinProgress"
	case stageWaitForSignatures:
		return "waitForSignatures"
	case stageTransactionFinalized:
		return "transactionFinalized"
	default:
		return fmt.Sprintf("stage(%d)", s)
	}
}
