package requestdata

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	iotago "github.com/iotaledger/iota.go"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type RequestDataType byte

const (
	RequestDataTypeUnknown = RequestDataType(iota)
	RequestDataTypeOffLedger
	RequestDataTypeUTXOSimple
	RequestDataTypeUTXOAlias
	RequestDataTypeUTXOExtended
	RequestDataTypeUTXONFT
	RequestDataTypeUTXOFoundry
	RequestDataTypeUTXOUnknown
)

var requestDataTypes = map[RequestDataType]string{
	RequestDataTypeUnknown:      "(wrong)",
	RequestDataTypeOffLedger:    "Off-ledger",
	RequestDataTypeUTXOSimple:   "SimpleUTXO",
	RequestDataTypeUTXOAlias:    "AliasUTXO",
	RequestDataTypeUTXOExtended: "ExtendedUTXO",
	RequestDataTypeUTXONFT:      "NTF-UTXO",
	RequestDataTypeUTXOFoundry:  "FoundryUTXO",
	RequestDataTypeUTXOUnknown:  "UnknownUTXO",
}

func (t RequestDataType) String() string {
	ret, ok := requestDataTypes[t]
	if ok {
		return ret
	}
	return "(wrong))"
}

type RequestNew interface {
	ID() iscp.RequestID
	Params() (dict.Dict, bool)
	SenderAccount() *iscp.AgentID
	SenderAddress() ledgerstate.Address
	Target() (iscp.Hname, iscp.Hname)
	Assets() (uint64, iotago.NativeTokens)
	GasBudget() int64
}

type RequestDataOptions interface {
	Timelock() (TimelockOptions, bool)
	Expiry() (ExpiryOptions, bool)
	ReturnAmount() (ReturnAmountOptions, bool)
	SwapOption() (SwapOptions, bool)
}

type TimelockOptions interface {
}

type ExpiryOptions interface {
}

type ReturnAmountOptions interface {
}

type SwapOptions interface {
}

// RequestData wraps any data which can be treated as a request under one interface
type RequestData interface {
	Type() RequestDataType

	Request() RequestNew
	Bytes() []byte
	String() string
	Unwrap() unwrap
	Options() RequestDataOptions
}

type unwrap interface {
	OffLedger() *OffLedger
	UTXO() unwrapUTXO
}

type unwrapUTXO interface {
	Simple() simpleOutput
	Alias() aliasOutput
	Extended() extendedOutput
	NFT() nftOutput
	Foundry() foundryOutput
	Unknown() unknownOutput
}

// placeholders
type simpleOutput struct{}
type aliasOutput struct{}
type extendedOutput struct{}
type nftOutput struct{}
type foundryOutput struct{}
type unknownOutput struct{}
