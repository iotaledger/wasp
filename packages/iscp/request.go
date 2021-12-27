// Wrapping interfaces for the request
// see also https://hackmd.io/@Evaldas/r1-L2UcDF and https://hackmd.io/@Evaldas/ryFK3Qr8Y and
package iscp

import (
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func UTXOInputFromMarshalUtil(mu *marshalutil.MarshalUtil) (*iotago.UTXOInput, error) {
	data, err := mu.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return nil, err
	}
	id, err := DecodeOutputID(data)
	if err != nil {
		return nil, err
	}
	return id.UTXOInput(), nil
}

func UTXOInputToMarshalUtil(id *iotago.UTXOInput, mu *marshalutil.MarshalUtil) {
	mu.WriteBytes(EncodeOutputID(id.ID()))
}

// RequestData wraps any data which can be potentially be interpreted as a request
type RequestData interface {
	Request

	IsOffLedger() bool
	AsOffLedger() AsOffLedger
	AsOnLedger() AsOnLedger

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
	SenderAccount() *AgentID // returns CommonAccount if sender address is not available
	SenderAddress() iotago.Address
	CallTarget() CallTarget
	TargetAddress() iotago.Address // TODO implement properly. Targte depends on time assumptions and AsUTXO type
	Assets() *Assets               // attached assets for the AsUTXO request, nil for off-ledger. All goes to sender
	Transfer() *Assets             // transfer of assets to the smart contract. Debited from sender account
	GasBudget() uint64
}

type Features interface {
	TimeLock() *TimeData
	Expiry() (*TimeData, iotago.Address) // return expiry time data and sender address or nil, nil if does not exist
	ReturnAmount() (uint64, bool)
}

type AsOffLedger interface {
	Nonce() uint64
}

type AsOnLedger interface {
	Output() iotago.Output
	IsInternalUTXO(*ChainID) bool
	UTXOInput() iotago.UTXOInput
	Features() Features
}

type ReturnAmountOptions interface {
	ReturnTo() iotago.Address
	Amount() uint64
}

func TakeRequestIDs(reqs ...RequestData) []RequestID {
	ret := make([]RequestID, len(reqs))
	for i := range reqs {
		ret[i] = reqs[i].ID()
	}
	return ret
}
