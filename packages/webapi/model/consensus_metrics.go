package model

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
)

type ConsensusWorkflowStatus struct {
	FlagStateReceived        bool `swagger:"desc(Shows if state output is received in current consensus iteration)"`
	FlagBatchProposalSent    bool `swagger:"desc(Shows if batch proposal is sent out in current consensus iteration)"`
	FlagConsensusBatchKnown  bool `swagger:"desc(Shows if consensus on batch is reached and known in current consensus iteration)"`
	FlagVMStarted            bool `swagger:"desc(Shows if virtual machine is started in current consensus iteration)"`
	FlagVMResultSigned       bool `swagger:"desc(Shows if virtual machine has returned its results in current consensus iteration)"`
	FlagTransactionFinalized bool `swagger:"desc(Shows if consensus on transaction is reached in current consensus iteration)"`
	FlagTransactionPosted    bool `swagger:"desc(Shows if transaction is posted to L1 in current consensus iteration)"`
	FlagTransactionSeen      bool `swagger:"desc(Shows if L1 reported that it has seen the transaction of current consensus iteration)"`
	FlagInProgress           bool `swagger:"desc(Shows if consensus algorithm is still not completed in current consensus iteration)"`

	TimeBatchProposalSent    time.Time `swagger:"desc(Shows when batch proposal was last sent out in current consensus iteration)"`
	TimeConsensusBatchKnown  time.Time `swagger:"desc(Shows when ACS results of consensus on batch was last received in current consensus iteration)"`
	TimeVMStarted            time.Time `swagger:"desc(Shows when virtual machine was last started in current consensus iteration)"`
	TimeVMResultSigned       time.Time `swagger:"desc(Shows when virtual machine results were last received and signed in current consensus iteration)"`
	TimeTransactionFinalized time.Time `swagger:"desc(Shows when algorithm last noted that all the data for consensus on transaction had been received in current consensus iteration)"`
	TimeTransactionPosted    time.Time `swagger:"desc(Shows when transaction was last posted to L1 in current consensus iteration)"`
	TimeTransactionSeen      time.Time `swagger:"desc(Shows when algorithm last noted that transaction hadd been seen by L1 in current consensus iteration)"`
	TimeCompleted            time.Time `swagger:"desc(Shows when algorithm was last completed in current consensus iteration)"`
}

func NewConsensusWorkflowStatus(status chain.ConsensusWorkflowStatus) *ConsensusWorkflowStatus {
	return &ConsensusWorkflowStatus{
		FlagStateReceived:        status.IsStateReceived(),
		FlagBatchProposalSent:    status.IsBatchProposalSent(),
		FlagConsensusBatchKnown:  status.IsConsensusBatchKnown(),
		FlagVMStarted:            status.IsVMStarted(),
		FlagVMResultSigned:       status.IsVMResultSigned(),
		FlagTransactionFinalized: status.IsTransactionFinalized(),
		FlagTransactionPosted:    status.IsTransactionPosted(),
		FlagTransactionSeen:      status.IsTransactionSeen(),
		FlagInProgress:           status.IsInProgress(),

		TimeBatchProposalSent:    status.GetBatchProposalSentTime(),
		TimeConsensusBatchKnown:  status.GetConsensusBatchKnownTime(),
		TimeVMStarted:            status.GetVMStartedTime(),
		TimeVMResultSigned:       status.GetVMResultSignedTime(),
		TimeTransactionFinalized: status.GetTransactionFinalizedTime(),
		TimeTransactionPosted:    status.GetTransactionPostedTime(),
		TimeTransactionSeen:      status.GetTransactionSeenTime(),
		TimeCompleted:            status.GetCompletedTime(),
	}
}
