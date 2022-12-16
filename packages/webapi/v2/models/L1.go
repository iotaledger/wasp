package models

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type Output struct {
	OutputType iotago.OutputType
	Bytes      []byte
}

func OutputFromIotaGoOutput(output iotago.Output) *Output {
	if output == nil {
		return nil
	}

	bytes, _ := output.Serialize(serializer.DeSeriModeNoValidation, nil)
	return &Output{
		OutputType: output.Type(),
		Bytes:      bytes,
	}
}

type OnLedgerRequest struct {
	ID       string
	OutputID string
	Output   *Output
	Bytes    []byte
}

func OnLedgerRequestFromISC(request isc.OnLedgerRequest) *OnLedgerRequest {
	if request == nil {
		return nil
	}

	return &OnLedgerRequest{
		ID:       request.ID().String(),
		OutputID: request.ID().OutputID().ToHex(),
		Output:   OutputFromIotaGoOutput(request.Output()),
		Bytes:    request.Bytes(),
	}
}

type InOutput struct {
	OutputID string
	Output   *Output
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
	OutputID string
	Output   *Output
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
	StateIndex    uint32
	TransactionID string
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
	TransactionID string
	State         string
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

type Transaction struct {
	TransactionID string
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

func TransactionFromIotaGoTransactionID(transaction *iotago.TransactionID) *Transaction {
	if transaction == nil {
		return nil
	}

	return &Transaction{
		TransactionID: transaction.ToHex(),
	}
}
