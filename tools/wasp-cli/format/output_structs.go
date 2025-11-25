package format

import "github.com/iotaledger/wasp/v2/tools/wasp-cli/log"

// ChainReceiptAsset mirrors the asset snippet included in chain receipt logs.
type ChainReceiptAsset struct {
	CoinType string `json:"coin_type"`
	Balance  string `json:"balance"`
}

// ChainReceiptOutput documents the JSON payload emitted when a chain receipt is logged.
//
// It contains both the formatter metadata (type/status/timestamp) and the receipt-specific
// fields so other packages (tests, tooling) can unmarshal CLI output without relying on
// map[string]interface{}.
type ChainReceiptOutput struct {
	Type               string              `json:"type,omitempty"`
	Status             string              `json:"status,omitempty"`
	Timestamp          string              `json:"timestamp,omitempty"`
	RequestID          string              `json:"request_id"`
	Kind               string              `json:"kind"`
	Sender             string              `json:"sender"`
	ContractHName      string              `json:"contract_hname"`
	FunctionHName      string              `json:"function_hname"`
	ParamsHex          []string            `json:"params_hex"`
	ArgumentsRaw       interface{}         `json:"arguments_raw"`
	DecodedKnownParams []log.TreeItem      `json:"decoded_known_params"`
	Error              string              `json:"error"`
	GasBudget          string              `json:"gas_budget"`
	GasBurned          string              `json:"gas_burned"`
	GasFeeCharged      string              `json:"gas_fee_charged"`
	StorageDeposit     string              `json:"storage_deposit"`
	Assets             []ChainReceiptAsset `json:"assets"`
	Index              *int                `json:"index,omitempty"`
}

// ToMap converts the struct into the map representation expected by FormatSuccess.
func (c ChainReceiptOutput) ToMap() map[string]interface{} {
	data := map[string]interface{}{
		"request_id":           c.RequestID,
		"kind":                 c.Kind,
		"sender":               c.Sender,
		"contract_hname":       c.ContractHName,
		"function_hname":       c.FunctionHName,
		"params_hex":           c.ParamsHex,
		"arguments_raw":        c.ArgumentsRaw,
		"decoded_known_params": c.DecodedKnownParams,
		"error":                c.Error,
		"gas_budget":           c.GasBudget,
		"gas_burned":           c.GasBurned,
		"gas_fee_charged":      c.GasFeeCharged,
		"storage_deposit":      c.StorageDeposit,
		"assets":               c.Assets,
	}
	if c.Index != nil {
		data["index"] = *c.Index
	}
	return data
}

// ChainBalanceCoin represents a single coin balance entry.
type ChainBalanceCoin struct {
	Token  string `json:"token"`
	Amount string `json:"amount"`
}

// ChainBalanceOutput documents the JSON payload emitted when a chain balance is logged.
//
// It contains both the formatter metadata (type/status/timestamp) and the balance-specific
// fields so other packages (tests, tooling) can unmarshal CLI output without relying on
// map[string]interface{}.
type ChainBalanceOutput struct {
	Type      string             `json:"type,omitempty"`
	Status    string             `json:"status,omitempty"`
	Timestamp string             `json:"timestamp,omitempty"`
	Coins     []ChainBalanceCoin `json:"coins"`
}

// ToMap converts the struct into the map representation expected by FormatSuccess.
func (c ChainBalanceOutput) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"coins": c.Coins,
	}
}
