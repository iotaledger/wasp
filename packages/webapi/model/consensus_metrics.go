package model

import (
	"time"
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

	CurrentStateIndex uint32 `swagger:"desc(Shows current state index of the consensus)"`
}

type ConsensusPipeMetrics struct {
	EventStateTransitionMsgPipeSize int
	EventPeerLogIndexMsgPipeSize    int
	EventACSMsgPipeSize             int
	EventVMResultMsgPipeSize        int
	EventTimerMsgPipeSize           int
}
