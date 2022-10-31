package models

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OffLedgerRequest struct {
	ChainID string `swagger:"desc(The chain id)"`

	Request string `swagger:"desc(Offledger Request (base64))"`
}

type ContractCallViewRequest struct {
	ChainID string `swagger:"desc(The chain id)"`

	ContractName  string
	ContractHName isc.Hname

	FunctionName  string
	FunctionHName isc.Hname

	Arguments dict.Dict
}
