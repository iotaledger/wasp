package model

import (
	"time"
)

type ConsensusWorkflowStatus struct {
	FlagStateReceived        bool `json:"flagStateReceived" swagger:"desc(Shows if state output is received in current consensus iteration)"`
	FlagBatchProposalSent    bool `json:"flagBatchProposalSent" swagger:"desc(Shows if batch proposal is sent out in current consensus iteration)"`
	FlagConsensusBatchKnown  bool `json:"flagConsensusBatchKnown" swagger:"desc(Shows if consensus on batch is reached and known in current consensus iteration)"`
	FlagVMStarted            bool `json:"flagVMStarted" swagger:"desc(Shows if virtual machine is started in current consensus iteration)"`
	FlagVMResultSigned       bool `json:"flagVMResultSigned" swagger:"desc(Shows if virtual machine has returned its results in current consensus iteration)"`
	FlagTransactionFinalized bool `json:"flagTransactionFinalized" swagger:"desc(Shows if consensus on transaction is reached in current consensus iteration)"`
	FlagTransactionPosted    bool `json:"flagTransactionPosted" swagger:"desc(Shows if transaction is posted to L1 in current consensus iteration)"`
	FlagTransactionSeen      bool `json:"flagTransactionSeen" swagger:"desc(Shows if L1 reported that it has seen the transaction of current consensus iteration)"`
	FlagInProgress           bool `json:"flagInProgress" swagger:"desc(Shows if consensus algorithm is still not completed in current consensus iteration)"`

	TimeBatchProposalSent    time.Time `json:"timeBatchProposalSent" swagger:"desc(Shows when batch proposal was last sent out in current consensus iteration)"`
	TimeConsensusBatchKnown  time.Time `json:"timeConsensusBatchKnown" swagger:"desc(Shows when ACS results of consensus on batch was last received in current consensus iteration)"`
	TimeVMStarted            time.Time `json:"timeVMStarted" swagger:"desc(Shows when virtual machine was last started in current consensus iteration)"`
	TimeVMResultSigned       time.Time `json:"timeVMResultSigned" swagger:"desc(Shows when virtual machine results were last received and signed in current consensus iteration)"`
	TimeTransactionFinalized time.Time `json:"timeTransactionFinalized" swagger:"desc(Shows when algorithm last noted that all the data for consensus on transaction had been received in current consensus iteration)"`
	TimeTransactionPosted    time.Time `json:"timeTransactionPosted" swagger:"desc(Shows when transaction was last posted to L1 in current consensus iteration)"`
	TimeTransactionSeen      time.Time `json:"timeTransactionSeen" swagger:"desc(Shows when algorithm last noted that transaction hadd been seen by L1 in current consensus iteration)"`
	TimeCompleted            time.Time `json:"timeCompleted" swagger:"desc(Shows when algorithm was last completed in current consensus iteration)"`

	CurrentStateIndex uint32 `json:"currentStateIndex" swagger:"desc(Shows current state index of the consensus)"`
}

type ConsensusPipeMetrics struct {
	EventStateTransitionMsgPipeSize int
	EventPeerLogIndexMsgPipeSize    int
	EventACSMsgPipeSize             int
	EventVMResultMsgPipeSize        int
	EventTimerMsgPipeSize           int
}
