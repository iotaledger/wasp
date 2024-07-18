package isc

import (
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type onLedgerRequestData struct {
	requestID sui.ObjectID

	requestMetadata *RequestMetadata
}

var (
	_ Request         = new(onLedgerRequestData)
	_ OnLedgerRequest = new(onLedgerRequestData)
	_ Calldata        = new(onLedgerRequestData)
)

func (req *onLedgerRequestData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOnLedger))
	rr.ReadN(req.requestID[:])
	outputData := rr.ReadBytes()
	if rr.Err != nil {
		return rr.Err
	}
	req.request, rr.Err = util.OutputFromBytes(outputData)
	if rr.Err != nil {
		return rr.Err
	}
	return req.readFromUTXO(req.request, req.requestID)
}

func (req *onLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(requestKindOnLedger))
	ww.WriteN(req.requestID[:])
	if ww.Err != nil {
		return ww.Err
	}
	outputData, err := req.request.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	ww.Err = err
	ww.WriteBytes(outputData)
	return ww.Err
}

func (req *onLedgerRequestData) Allowance() *Assets {
	if req.requestMetadata == nil {
		return NewEmptyAssets()
	}
	return req.requestMetadata.Allowance
}

func (req *onLedgerRequestData) Assets() *Assets {
	amount := req.output.Deposit()
	// FIXME work on the SUI version
	// tokens := req.output.NativeTokenList()
	ret := NewAssets(new(big.Int).SetUint64(amount), nil)
	return ret
}

func (req *onLedgerRequestData) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *onLedgerRequestData) Message() Message {
	if req.requestMetadata == nil {
		return Message{}
	}
	return req.requestMetadata.Message
}

func (req *onLedgerRequestData) Clone() OnLedgerRequest {
	outputID := iotago.OutputID{}
	copy(outputID[:], req.requestID[:])

	ret := &onLedgerRequestData{
		requestID:        outputID,
		request:          req.request.Clone(),
		featureBlocks:    req.featureBlocks.Clone(),
		unlockConditions: util.CloneMap(req.unlockConditions),
	}
	if req.requestMetadata != nil {
		ret.requestMetadata = req.requestMetadata.Clone()
	}
	return ret
}

func (req *onLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	if req.requestMetadata == nil {
		return 0, false
	}
	return req.requestMetadata.GasBudget, false
}

func (req *onLedgerRequestData) ID() RequestID {
	return RequestID(req.requestID)
}

func (req *onLedgerRequestData) IsOffLedger() bool {
	return false
}

func (req *onLedgerRequestData) RequestID() sui.ObjectID {
	return req.requestID
}

func (req *onLedgerRequestData) SenderAccount() AgentID {
	sender := req.senderAddress()
	if sender == nil {
		return nil
	}
	if req.requestMetadata != nil && !req.requestMetadata.SenderContract.Empty() {
		// if sender.Type() == iotago.AddressAlias {	// TODO: is it needed?
		chainID := ChainIDFromAddress(sender)
		return req.requestMetadata.SenderContract.AgentID(chainID)
		//}
	}
	return NewAgentID(sender)
}

func (req *onLedgerRequestData) senderAddress() *cryptolib.Address {
	senderBlock := req.featureBlocks.SenderFeature()
	if senderBlock == nil {
		return nil
	}

	return cryptolib.NewAddressFromIotago(senderBlock.Address)
}

func (req *onLedgerRequestData) String() string {
	metadata := req.requestMetadata
	if metadata == nil {
		return "onledger request without metadata"
	}
	return fmt.Sprintf("onLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, GasBudget: %d }",
		req.ID().String(),
		metadata.SenderContract.String(),
		metadata.Message.Target.Contract.String(),
		metadata.Message.Target.EntryPoint.String(),
		metadata.Message.Params.String(),
		metadata.GasBudget,
	)
}

func (req *onLedgerRequestData) TargetAddress() *cryptolib.Address {
	// TODO: refactor me: (?) Is TargetAddress still needed? It will always be the ChainID anyway, I think.
	/*switch out := req.request.(type) {
	case *iotago.BasicOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.FoundryOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.NFTOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.AliasOutput:
		return cryptolib.NewAddressFromIotago(out.AliasID.ToAddress())
	default:
		panic("onLedgerRequestData:TargetAddress implement me")
	}*/

	return req.request.Anchor
}

func (req *onLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}

// region RetryOnLedgerRequest //////////////////////////////////////////////////////////////////

type RetryOnLedgerRequest struct {
	OnLedgerRequest
	retryOutputID sui.ObjectID
}

func NewRetryOnLedgerRequest(req OnLedgerRequest, retryOutput sui.ObjectID) *RetryOnLedgerRequest {
	return &RetryOnLedgerRequest{
		OnLedgerRequest: req,
		retryOutputID:   retryOutput,
	}
}

func (r *RetryOnLedgerRequest) RetryOutputID() sui.ObjectID {
	return r.retryOutputID
}

func (r *RetryOnLedgerRequest) SetRetryOutputID(oid sui.ObjectID) {
	r.retryOutputID = oid
}
