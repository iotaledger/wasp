package models

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedgerRequest struct {
	ChainID string `swagger:"desc(The chain id)"`

	Request string `swagger:"desc(Offledger Request (Hex))"`
}

type ContractCallViewRequest struct {
	ChainID string `swagger:"desc(The chain id)"`

	ContractName  string    `swagger:"desc(The contract name)"`
	ContractHName isc.Hname `swagger:"desc(The contract name as HName)"`

	FunctionName  string    `swagger:"desc(The function name)"`
	FunctionHName isc.Hname `swagger:"desc(The function name as HName)"`

	Arguments dict.JSONDict `swagger:"desc(Encoded arguments to be passed to the function)"`
}
