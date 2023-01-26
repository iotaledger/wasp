package models

import (
	"github.com/iotaledger/hive.go/core/typeutils"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type Output struct {
	OutputType iotago.OutputType `json:"outputType" swagger:"desc(The output type),required"`
	Raw        string            `json:"raw" swagger:"desc(The raw data of the output (Hex)),required"`
}

func OutputFromIotaGoOutput(output iotago.Output) *Output {
	if typeutils.IsInterfaceNil(output) {
		return nil
	}

	bytes, _ := output.Serialize(serializer.DeSeriModeNoValidation, nil)
	return &Output{
		OutputType: output.Type(),
		Raw:        iotago.EncodeHex(bytes),
	}
}

type OnLedgerRequest struct {
	ID       string  `json:"id" swagger:"desc(The request ID),required"`
	OutputID string  `json:"outputId" swagger:"desc(The output ID),required"`
	Output   *Output `json:"output" swagger:"desc(The parsed output),required"`
	Raw      string  `json:"raw" swagger:"desc(The raw data of the request (Hex)),required"`
}

func OnLedgerRequestFromISC(request isc.OnLedgerRequest) *OnLedgerRequest {
	if typeutils.IsInterfaceNil(request) {
		return nil
	}

	return &OnLedgerRequest{
		ID:       request.ID().String(),
		OutputID: request.ID().OutputID().ToHex(),
		Output:   OutputFromIotaGoOutput(request.Output()),
		Raw:      iotago.EncodeHex(request.Bytes()),
	}
}

type InOutput struct {
	OutputID string  `json:"outputId" swagger:"desc(The output ID),required"`
	Output   *Output `json:"output" swagger:"desc(The parsed output),required"`
}

func InOutputFromISCInOutput(output *nodeconnmetrics.InOutput) *InOutput {
	if output == nil {
		return nil
	}

	return &InOutput{
		OutputID: output.OutputID.ToHex(),
		Output:   OutputFromIotaGoOutput(output.Output),
	}
}

type InStateOutput struct {
	OutputID string  `json:"outputId" swagger:"desc(The output ID),required"`
	Output   *Output `json:"output" swagger:"desc(The parsed output),required"`
}

func InStateOutputFromISCInStateOutput(output *nodeconnmetrics.InStateOutput) *InStateOutput {
	if output == nil {
		return nil
	}

	return &InStateOutput{
		OutputID: output.OutputID.ToHex(),
		Output:   OutputFromIotaGoOutput(output.Output),
	}
}

type StateTransaction struct {
	StateIndex    uint32 `json:"stateIndex" swagger:"desc(The state index),required,min(1)"`
	TransactionID string `json:"txId" swagger:"desc(The transaction ID),required"`
}

func StateTransactionFromISCStateTransaction(transaction *nodeconnmetrics.StateTransaction) *StateTransaction {
	if transaction == nil {
		return nil
	}

	txID, _ := transaction.Transaction.ID()

	return &StateTransaction{
		StateIndex:    transaction.StateIndex,
		TransactionID: txID.ToHex(),
	}
}

type TxInclusionStateMsg struct {
	TransactionID string `json:"txId" swagger:"desc(The transaction ID),required"`
	State         string `json:"state" swagger:"desc(The inclusion state),required"`
}

func TxInclusionStateMsgFromISCTxInclusionStateMsg(inclusionState *nodeconnmetrics.TxInclusionStateMsg) *TxInclusionStateMsg {
	if inclusionState == nil {
		return nil
	}

	return &TxInclusionStateMsg{
		State:         inclusionState.State,
		TransactionID: inclusionState.TxID.ToHex(),
	}
}

type MilestoneInfo struct {
	Index       uint32 `json:"index"`
	Timestamp   uint32 `json:"timestamp"`
	MilestoneID string `json:"milestoneId"`
}

func MilestoneFromIotaGoMilestone(milestone *nodeclient.MilestoneInfo) *MilestoneInfo {
	if milestone == nil {
		return &MilestoneInfo{}
	}

	return &MilestoneInfo{
		Index:       milestone.Index,
		MilestoneID: milestone.MilestoneID,
		Timestamp:   milestone.Timestamp,
	}
}

type Transaction struct {
	TransactionID string `json:"txId" swagger:"desc(The transaction ID),required"`
}

type OutputID struct {
	OutputID string `json:"outputId" swagger:"desc(The output ID),required"`
}

func TransactionFromIotaGoTransaction(transaction *iotago.Transaction) *Transaction {
	if transaction == nil {
		return nil
	}

	txID, _ := transaction.ID()

	return &Transaction{
		TransactionID: txID.ToHex(),
	}
}

func TransactionFromIotaGoTransactionID(txID *iotago.TransactionID) *Transaction {
	if txID == nil {
		return nil
	}

	return &Transaction{
		TransactionID: txID.ToHex(),
	}
}

func OutputIDFromIotaGoOutputID(outputID iotago.OutputID) *OutputID {
	return &OutputID{
		OutputID: outputID.ToHex(),
	}
}
