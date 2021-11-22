// Wrapping interfaces for the request
// see also https://hackmd.io/@Evaldas/r1-L2UcDF and https://hackmd.io/@Evaldas/ryFK3Qr8Y and
package iscp

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// UTXOMetaData is coming together with UTXO from L1
// It is a part of each implementation of RequestData
type UTXOMetaData struct {
	UTXOInput          iotago.UTXOInput
	MilestoneIndex     uint32
	MilestoneTimestamp time.Time
}

func (u *UTXOMetaData) Bytes() []byte {
	panic("not implemented")
}

func UTXOMetaDataFromBytes() (*UTXOMetaData, error) {
	panic("not implemented") // TODO
}

// RequestData wraps any data which can be potentially be interpreted as a request
type RequestData interface {
	Request
	IsOffLedger() bool
	Unwrap() unwrap

	Bytes() []byte
	String() string
}

type TimeData struct {
	MilestoneIndex uint32
	Time           time.Time
}

type NFT struct {
	NFTID       iotago.NFTID
	NFTMetadata []byte
}

type Request interface {
	ID() RequestID
	Params() dict.Dict
	SenderAccount() *AgentID // returns CommonAccount if sender address is ot available
	SenderAddress() iotago.Address
	Target() RequestTarget
	Assets() *Assets   // attached assets for the UTXO request, nil for off-ledger. All goes to sender
	Transfer() *Assets // transfer of assets to the smart contract. Debited from sender account
	GasBudget() uint64
}

type Features interface {
	TimeLock() *TimeData
	Expiry() (*TimeData, iotago.Address) // return expiry time data and sender address or nil, nil if does not exist
	ReturnAmount() (uint64, bool)
}

type unwrap interface {
	OffLedger() *OffLedgerRequestData
	UTXO() unwrapUTXO
}

type unwrapUTXO interface {
	Output() iotago.Output
	Metadata() *UTXOMetaData
	Features() Features
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
