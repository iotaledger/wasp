package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/hive.go/objectstorage/typeutils"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
)

type OnLedgerRequest struct {
	ID  string `json:"id" swagger:"desc(The request ID),required"`
	Raw string `json:"raw" swagger:"desc(The raw data of the request (Hex)),required"`
}

func OnLedgerRequestFromISC(request isc.OnLedgerRequest) *OnLedgerRequest {
	if typeutils.IsInterfaceNil(request) {
		return nil
	}

	return &OnLedgerRequest{
		ID:  request.ID().Short(),
		Raw: request.String(),
	}
}

type StateAnchor struct {
	StateIndex    uint32 `json:"stateIndex" swagger:"desc(The state index),required,min(1)"`
	StateMetadata string `json:"stateMetadata" swagger:"desc(The state metadata),required"`
	Raw           string `json:"raw" swagger:"desc(The raw data of the anchor (Hex)),required"`
}

type StateTransaction struct {
	StateIndex        uint32 `json:"stateIndex" swagger:"desc(The state index),required,min(1)"`
	TransactionDigest string `json:"txDigest" swagger:"desc(The transaction Digest),required"`
}

func StateAnchorFromISCStateAnchor(anchor *metrics.StateAnchor) *StateAnchor {
	if anchor == nil {
		return nil
	}

	b, err := bcs.Marshal(anchor)
	if err != nil {
		return nil
	}

	return &StateAnchor{
		StateIndex:    anchor.StateIndex,
		StateMetadata: anchor.StateMetadata,
		Raw:           hexutil.Encode(b),
	}
}

func StateTransactionFromISCStateTransaction(transaction *metrics.StateTransaction) *StateTransaction {
	if transaction == nil {
		return nil
	}

	return &StateTransaction{
		StateIndex:        transaction.StateIndex,
		TransactionDigest: transaction.TransactionDigest.String(),
	}
}
