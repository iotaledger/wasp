package models

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedgerRequest struct {
	ChainID string `json:"chainId" swagger:"desc(The chain id),required"`
	Request string `json:"request" swagger:"desc(Offledger Request (Hex)),required"`
}

type ContractCallViewRequest struct {
	ChainID       string        `json:"chainId" swagger:"desc(The chain id),required"`
	ContractName  string        `json:"contractName" swagger:"desc(The contract name),required"`
	ContractHName string        `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionName  string        `json:"functionName" swagger:"desc(The function name),required"`
	FunctionHName string        `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
	Arguments     dict.JSONDict `json:"arguments" swagger:"desc(Encoded arguments to be passed to the function),required"`
}
