package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/hive.go/objectstorage/typeutils"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
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

// TODO should be removed in the future
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

type L1Params struct {
	Protocol  *Protocol     `json:"protocol" swagger:"required"`
	BaseToken *IotaCoinInfo `json:"baseToken" swagger:"required"`
}

func MapL1ParamsResponse(l1Params *parameters.L1Params) L1Params {
	return L1Params{
		Protocol: &Protocol{
			Epoch:                 l1Params.Protocol.Epoch.String(),
			ProtocolVersion:       l1Params.Protocol.ProtocolVersion.String(),
			SystemStateVersion:    l1Params.Protocol.SystemStateVersion.String(),
			ReferenceGasPrice:     l1Params.Protocol.ReferenceGasPrice.String(),
			EpochStartTimestampMs: l1Params.Protocol.EpochStartTimestampMs.String(),
			EpochDurationMs:       l1Params.Protocol.EpochDurationMs.String(),
		},
		BaseToken: &IotaCoinInfo{
			CoinType:    l1Params.BaseToken.CoinType.String(),
			Name:        l1Params.BaseToken.Name,
			Symbol:      l1Params.BaseToken.Symbol,
			Description: l1Params.BaseToken.Description,
			IconURL:     l1Params.BaseToken.IconURL,
			Decimals:    l1Params.BaseToken.Decimals,
			TotalSupply: l1Params.BaseToken.TotalSupply.String(),
		},
	}
}

type Protocol struct {
	Epoch                 string `json:"epoch" swagger:"required"`
	ProtocolVersion       string `json:"protocol_version" swagger:"required"`
	SystemStateVersion    string `json:"system_state_version" swagger:"required"`
	ReferenceGasPrice     string `json:"reference_gas_price" swagger:"required"`
	EpochStartTimestampMs string `json:"epoch_start_timestamp_ms" swagger:"required"`
	EpochDurationMs       string `json:"epoch_duration_ms" swagger:"required"`
}

type IotaCoinInfo struct {
	CoinType    string `json:"coinType" swagger:"desc(BaseToken's Cointype),required"`
	Name        string `json:"name" swagger:"desc(The base token name),required"`
	Symbol      string `json:"tickerSymbol" swagger:"desc(The ticker symbol),required"`
	Description string `json:"description" swagger:"desc(The token description),required"`
	IconURL     string `json:"iconUrl" swagger:"desc(The icon URL),required"`
	Decimals    uint8  `json:"decimals" swagger:"desc(The token decimals),required"`
	TotalSupply string `json:"totalSupply" swagger:"desc(The total supply of BaseToken, a string for uint64),required"`
}
