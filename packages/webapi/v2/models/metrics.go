package models

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type MetricItem[T interface{}] struct {
	Messages    uint32
	Timestamp   time.Time
	LastMessage T
}

/*
ChainMetrics
Echo Swagger does not support generics such as MetricItem[Foo]
Creating separate types works instead.
To not create a mapper for each type, the actual service remains using MetricItem[Foo] but this model here is presented to the docs.
This can be removed if we change to swag/echo-swagger
*/
type (
	AliasOutputMetricItem         MetricItem[*Output]
	OnLedgerRequestMetricItem     MetricItem[*OnLedgerRequest]
	InOutputMetricItem            MetricItem[*InOutput]
	InStateOutputMetricItem       MetricItem[*InStateOutput]
	TxInclusionStateMsgMetricItem MetricItem[*TxInclusionStateMsg]
	TransactionMetricItem         MetricItem[*Transaction]
	TransactionIDMetricItem       MetricItem[*Transaction]
	UTXOInputMetricItem           MetricItem[iotago.OutputID]
	InterfaceMetricItem           MetricItem[interface{}]
)

type ChainMetrics struct {
	InAliasOutput                   AliasOutputMetricItem
	InOnLedgerRequest               OnLedgerRequestMetricItem
	InOutput                        InOutputMetricItem
	InStateOutput                   InStateOutputMetricItem
	InTxInclusionState              TxInclusionStateMsgMetricItem
	OutPublishGovernanceTransaction TransactionMetricItem
	OutPullLatestOutput             InterfaceMetricItem
	OutPullOutputByID               UTXOInputMetricItem
	OutPullTxInclusionState         TransactionIDMetricItem
}

func MapMetricItem[T any, G any](metrics *dto.MetricItem[G], value T) MetricItem[T] {
	return MetricItem[T]{
		Messages:    metrics.Messages,
		Timestamp:   metrics.Timestamp,
		LastMessage: value,
	}
}

func MapChainMetrics(metrics *dto.ChainMetrics) *ChainMetrics {
	return &ChainMetrics{
		InAliasOutput:                   AliasOutputMetricItem(MapMetricItem(metrics.InAliasOutput, OutputFromIotaGoOutput(metrics.InAliasOutput.LastMessage))),
		InOutput:                        InOutputMetricItem(MapMetricItem(metrics.InOutput, InOutputFromISCInOutput(metrics.InOutput.LastMessage))),
		InTxInclusionState:              TxInclusionStateMsgMetricItem(MapMetricItem(metrics.InTxInclusionState, TxInclusionStateMsgFromISCTxInclusionStateMsg(metrics.InTxInclusionState.LastMessage))),
		InOnLedgerRequest:               OnLedgerRequestMetricItem(MapMetricItem(metrics.InOnLedgerRequest, OnLedgerRequestFromISC(metrics.InOnLedgerRequest.LastMessage))),
		OutPullOutputByID:               UTXOInputMetricItem(MapMetricItem(metrics.OutPullOutputByID, metrics.OutPullOutputByID.LastMessage)),
		OutPullTxInclusionState:         TransactionIDMetricItem(MapMetricItem(metrics.OutPullTxInclusionState, TransactionFromIotaGoTransactionID(&metrics.OutPullTxInclusionState.LastMessage))),
		OutPullLatestOutput:             InterfaceMetricItem(MapMetricItem(metrics.OutPullLatestOutput, metrics.OutPullLatestOutput.LastMessage)),
		InStateOutput:                   InStateOutputMetricItem(MapMetricItem(metrics.InStateOutput, InStateOutputFromISCInStateOutput(metrics.InStateOutput.LastMessage))),
		OutPublishGovernanceTransaction: TransactionMetricItem(MapMetricItem(metrics.OutPublishGovernanceTransaction, TransactionFromIotaGoTransaction(metrics.OutPublishGovernanceTransaction.LastMessage))),
	}
}

type ConsensusWorkflowMetrics struct {
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

func MapConsensusWorkflowStatus(status chain.ConsensusWorkflowStatus) *ConsensusWorkflowMetrics {
	return &ConsensusWorkflowMetrics{
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

		CurrentStateIndex: status.GetCurrentStateIndex(),
	}
}

type ConsensusPipeMetrics struct {
	EventStateTransitionMsgPipeSize int
	EventPeerLogIndexMsgPipeSize    int
	EventACSMsgPipeSize             int
	EventVMResultMsgPipeSize        int
	EventTimerMsgPipeSize           int
}

func MapConsensusPipeMetrics(pipeMetrics chain.ConsensusPipeMetrics) *ConsensusPipeMetrics {
	return &ConsensusPipeMetrics{
		EventStateTransitionMsgPipeSize: pipeMetrics.GetEventStateTransitionMsgPipeSize(),
		EventPeerLogIndexMsgPipeSize:    pipeMetrics.GetEventPeerLogIndexMsgPipeSize(),
		EventACSMsgPipeSize:             pipeMetrics.GetEventACSMsgPipeSize(),
		EventVMResultMsgPipeSize:        pipeMetrics.GetEventVMResultMsgPipeSize(),
		EventTimerMsgPipeSize:           pipeMetrics.GetEventTimerMsgPipeSize(),
	}
}
