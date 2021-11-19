// Wrapping interfaces for the request
// see also https://hackmd.io/@Evaldas/r1-L2UcDF and https://hackmd.io/@Evaldas/ryFK3Qr8Y and
package iscp

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type TypeCode byte

const (
	TypeUnknown = TypeCode(iota)
	TypeOffLedger
	TypeAliasOutput
	TypeExtendedOutput
	TypeFoundryOutput
	TypeNFTOutput
	TypeUnknownOutput
)

var typeCodes = map[TypeCode]string{
	TypeUnknown:        "(wrong)",
	TypeOffLedger:      "Off-ledger",
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
	return "(wrong)"
}

// UTXOMetaData is coming together with UTXO from L1
// It is a part of each implementation of RequestData
type UTXOMetaData struct {
	UTXOInput          iotago.UTXOInput
	MilestoneIndex     uint32
	MilestoneTimestamp uint64
}

// RequestData wraps any data which can be potentially be interpreted as a request
type RequestData interface {
	Type() TypeCode

	Request() Request // nil if the RequestData cannot be interpreted as request, for example does not contain Sender
	TimeData() *TimeData

	Unwrap() unwrap
	Features() Features

	Bytes() []byte
	String() string
}

type TimeData struct {
	MilestoneIndex uint32
	Timestamp      uint64
}

type NFT struct {
	NFTID       iotago.NFTID
	NFTMetadata []byte
}

type Request interface {
	ID() RequestID
	Params() dict.Dict
	SenderAccount() *AgentID
	SenderAddress() iotago.Address
	Target() RequestTarget
	Assets() *Assets   // attached assets for the UTXO request, nil for off-ledger. All goes to sender
	Transfer() *Assets // transfer of assets to the smart contract. Debited from sender account
	GasBudget() int64
}

type Features interface {
	TimeLock() *TimeData
	Expiry() *TimeData
	ReturnAmount() (uint64, bool)
}

type unwrap interface {
	OffLedger() *OffLedger
	UTXO() iotago.Output
}

type ReturnAmountOptions interface {
	ReturnTo() iotago.Address
	Amount() uint64
}

type RequestTarget struct {
	Contract   Hname
	EntryPoint Hname
}

func NewRequestTarget(contract, entryPoint Hname) RequestTarget {
	return RequestTarget{
		Contract:   contract,
		EntryPoint: entryPoint,
	}
}

func TakeRequestIDs(reqs ...Request) []RequestID {
	ret := make([]RequestID, len(reqs))
	for i := range reqs {
		ret[i] = reqs[i].ID()
	}
	return ret
}

func (txm *UTXOMetaData) RequestID() RequestID {
	return RequestID(txm.UTXOInput)
}
