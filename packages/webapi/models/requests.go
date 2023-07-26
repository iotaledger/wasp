package models

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedgerRequest struct {
	ChainID string `json:"chainId" swagger:"desc(The chain id),required"`
	Request string `json:"request" swagger:"desc(Offledger Request (Hex)),required"`
}

type ContractCallViewRequest struct {
	ContractName  string        `json:"contractName" swagger:"desc(The contract name),required"`
	ContractHName string        `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionName  string        `json:"functionName" swagger:"desc(The function name),required"`
	FunctionHName string        `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
	Arguments     dict.JSONDict `json:"arguments" swagger:"desc(Encoded arguments to be passed to the function),required"`
	Block         string        `json:"block" swagger:"desc(block index or trie root to execute the view call in, latest block will be used if not specified)"`
}

type EstimateGasRequestOnledger struct {
	Output string `json:"outputBytes" swagger:"desc(Serialized Output (Hex)),required"`
}

type EstimateGasRequestOffledger struct {
	Request string `json:"requestBytes" swagger:"desc(Offledger Request (Hex)),required"`
}
