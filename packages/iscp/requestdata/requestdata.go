// Wrapping interfaces for the request
// see also https://hackmd.io/@Evaldas/r1-L2UcDF and https://hackmd.io/@Evaldas/ryFK3Qr8Y and
package requestdata

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type TypeCode byte

const (
	TypeUnknown = TypeCode(iota)
	TypeOffLedger
	TypeSimpleOutput
	TypeAliasOutput
	TypeExtendedOutput
	TypeFoundryOutput
	TypeNFTOutput
	TypeUnknownOutput
)

var typeCodes = map[TypeCode]string{
	TypeUnknown:        "(wrong)",
	TypeOffLedger:      "Off-ledger",
	TypeSimpleOutput:   "SimpleUTXO",
	TypeAliasOutput:    "AliasUTXO",
	TypeExtendedOutput: "ExtendedUTXO",
	TypeNFTOutput:      "NTF-UTXO",
	TypeFoundryOutput:  "FoundryUTXO",
	TypeUnknownOutput:  "UnknownUTXO",
}

func (t TypeCode) String() string {
	ret, ok := typeCodes[t]
	if ok {
		return ret
	}
	return "(wrong))"
}

// UTXOMetaData is coming together with UTXO from L1
// It is a part of each implementation of RequestData
type UTXOMetaData struct {
	UTXOInput          iotago.UTXOInput
	MilestoneIndex     uint32
	MilestoneTimestamp time.Time
}

// RequestData wraps any data which can be potentially be interpreted as a request
type RequestData interface {
	Type() TypeCode

	Request() Request
	TimeData() *TimeData

	MustUnwrap() unwrap
	Features() Features

	Bytes() []byte
	String() string
}

type TimeData struct {
	ConfirmingMilestoneIndex uint32
	ConfirmationTime         time.Time // should better be UnixNano ?
}

type Request interface {
	ID() RequestID
	Params() dict.Dict
	SenderAccount() *iscp.AgentID
	SenderAddress() iotago.Address
	Target() (iscp.Hname, iscp.Hname)
	Assets() (uint64, iotago.NativeTokens)
	GasBudget() int64
}

type Features interface {
	TimeLock() (TimeLockOptions, bool)
	Expiry() (ExpiryOptions, bool)
	ReturnAmount() (ReturnAmountOptions, bool)
	SwapOption() (SwapOptions, bool) // for the new swap
}

type unwrap interface {
	OffLedger() *OffLedger
	UTXO() unwrapUTXO
}

type unwrapUTXO interface {
	MetaData() UTXOMetaData
	Simple() *iotago.SimpleOutput
	Alias() *iotago.AliasOutput
	Extended() *iotago.ExtendedOutput
	NFT() *iotago.NFTOutput
	Foundry() *iotago.FoundryOutput
	Unknown() *placeholders.UnknownOutput
}

type TimeLockOptions interface {
	Deadline() (time.Time, bool)
	MilestoneIndex() (uint32, bool)
}

type ExpiryOptions interface {
	Deadline() time.Time
}

type ReturnAmountOptions interface {
	Amount() uint64
}

type SwapOptions interface {
	ExpiryOptions
	ReturnAmountOptions
}

func (txm *UTXOMetaData) RequestID() RequestID {
	return RequestID(txm.UTXOInput)
}
