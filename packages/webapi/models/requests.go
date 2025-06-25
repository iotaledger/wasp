package models

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/wasp/packages/isc"
)

type RequestJSON struct {
	AllowanceError string            `json:"allowanceError,omitempty"`
	Allowance      *AssetsJSON       `json:"allowance" swagger:"required"`
	CallTarget     CallTargetJSON    `json:"callTarget" swagger:"required"`
	Assets         AssetsJSON        `json:"assets" swagger:"required"`
	GasBudget      string            `json:"gasBudget,string" swagger:"required,desc(The gas budget (uint64 as string))"`
	IsEVM          bool              `json:"isEVM" swagger:"required"`
	IsOffLedger    bool              `json:"isOffLedger" swagger:"required"`
	Params         CallArgumentsJSON `json:"params" swagger:"required"`
	RequestID      string            `json:"requestId" swagger:"required"`
	SenderAccount  string            `json:"senderAccount" swagger:"required"`
}

func RequestToJSONObject(request isc.Request) RequestJSON {
	gasBudget, isEVM := request.GasBudget()
	msg := request.Message()

	r := RequestJSON{
		CallTarget:    callTargetToJSONObject(msg.Target),
		Assets:        AssetsToAssetsJSON(request.Assets()),
		GasBudget:     strconv.FormatUint(gasBudget, 10),
		IsEVM:         isEVM,
		IsOffLedger:   request.IsOffLedger(),
		Params:        ToCallArgumentsJSON(msg.Params),
		RequestID:     request.ID().String(),
		SenderAccount: request.SenderAccount().String(),
	}

	allowance, err := request.Allowance()
	if err != nil {
		r.AllowanceError = err.Error()
	}
	if allowance != nil {
		allowanceJSON := AssetsToAssetsJSON(allowance)
		r.Allowance = &allowanceJSON
	}

	return r
}

// ----------------------------------------------------------------------------

type CallTargetJSON struct {
	ContractHName string `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionHName string `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
}

func callTargetToJSONObject(target isc.CallTarget) CallTargetJSON {
	return CallTargetJSON{
		ContractHName: target.Contract.String(),
		FunctionHName: target.EntryPoint.String(),
	}
}

type CallArgumentsJSON []string

func ToCallArgumentsJSON(c isc.CallArguments) CallArgumentsJSON {
	callArgumentsJSON := make(CallArgumentsJSON, len(c))
	for i, v := range c {
		callArgumentsJSON[i] = hexutil.Encode(v)
	}
	return callArgumentsJSON
}

type CallResultsJSON CallArgumentsJSON

func (c CallResultsJSON) ToCallResults() (isc.CallArguments, error) {
	callResults := make(isc.CallResults, len(c))
	var err error
	for i, v := range c {
		callResults[i], err = hexutil.Decode(v)
		if err != nil {
			return nil, err
		}
	}
	return callResults, nil
}

func ToCallResultsJSON(c isc.CallResults) CallResultsJSON {
	return CallResultsJSON(ToCallArgumentsJSON(c))
}

type OffLedgerRequest struct {
	Request string `json:"request" swagger:"desc(Offledger Request (Hex)),required"`
}

type ContractCallViewRequest struct {
	ContractName  string   `json:"contractName" swagger:"desc(The contract name),required"`
	ContractHName string   `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionName  string   `json:"functionName" swagger:"desc(The function name),required"`
	FunctionHName string   `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
	Arguments     []string `json:"arguments" swagger:"desc(Encoded arguments to be passed to the function),required"`
	Block         string   `json:"block" swagger:"desc(block index or trie root to execute the view call in, latest block will be used if not specified)"`
}

type EstimateGasRequestOnledger struct {
	Output string `json:"outputBytes" swagger:"desc(Serialized Output (Hex)),required"`
}

type EstimateGasRequestOffledger struct {
	Request     string `json:"requestBytes" swagger:"desc(Offledger Request (Hex)),required"`
	FromAddress string `json:"fromAddress" swagger:"desc(The address to estimate gas for(Hex)),required"`
}
