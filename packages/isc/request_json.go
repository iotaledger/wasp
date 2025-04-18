package isc

import (
	"encoding/json"
	"strconv"
)

type RequestJSON struct {
	AllowanceError string            `json:"allowanceError,omitempty"`
	Allowance      *AssetsJSON       `json:"allowance" swagger:"required"`
	CallTarget     CallTargetJSON    `json:"callTarget" swagger:"required"`
	Assets         AssetsJSON        `json:"assets" swagger:"required"`
	GasBudget      string            `json:"gasBudget" swagger:"required,desc(The gas budget (uint64 as string))"`
	IsEVM          bool              `json:"isEVM" swagger:"required"`
	IsOffLedger    bool              `json:"isOffLedger" swagger:"required"`
	Params         CallArgumentsJSON `json:"params" swagger:"required"`
	RequestID      string            `json:"requestId" swagger:"required"`
	SenderAccount  string            `json:"senderAccount" swagger:"required"`
}

func RequestToJSONObject(request Request) RequestJSON {
	gasBudget, isEVM := request.GasBudget()
	msg := request.Message()

	r := RequestJSON{
		CallTarget:    callTargetToJSONObject(msg.Target),
		Assets:        AssetsToAssetsJSON(request.Assets()),
		GasBudget:     strconv.FormatUint(gasBudget, 10),
		IsEVM:         isEVM,
		IsOffLedger:   request.IsOffLedger(),
		Params:        msg.Params.ToCallArgumentsJSON(),
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

func RequestToJSON(req Request) ([]byte, error) {
	return json.Marshal(RequestToJSONObject(req))
}

// ----------------------------------------------------------------------------

type CallTargetJSON struct {
	ContractHName string `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionHName string `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
}

func callTargetToJSONObject(target CallTarget) CallTargetJSON {
	return CallTargetJSON{
		ContractHName: target.Contract.String(),
		FunctionHName: target.EntryPoint.String(),
	}
}
