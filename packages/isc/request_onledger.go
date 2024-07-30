package isc

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/types"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type onLedgerRequestData struct {
	requestID       sui.ObjectID
	senderAddress   *cryptolib.Address
	targetAddress   *cryptolib.Address
	assets          *Assets
	requestMetadata *RequestMetadata
}

var (
	_ Request         = new(onLedgerRequestData)
	_ OnLedgerRequest = new(onLedgerRequestData)
	_ Calldata        = new(onLedgerRequestData)
)

func OnLedgerFromRequest(request types.Request) (OnLedgerRequest, error) {
	r := &onLedgerRequestData{
		requestID:     request.ID,
		senderAddress: request.Sender,
		targetAddress: request.Anchor,
		requestMetadata: &RequestMetadata{
			SenderContract: ContractIdentity{},
			Message: Message{
				Target: CallTarget{
					Contract:   Hname(request.Message.Contract),
					EntryPoint: Hname(request.Message.Function),
				},
				Params: dict.New(), // TODO: set request.Message.Params (turn dict to list of args)
			},
			Allowance: NewEmptyAssets(),
			GasBudget: 0,
		},
	}

	return r, nil
}

func (req *onLedgerRequestData) Read(r io.Reader) error {
	var err error
	rr := rwutil.NewReader(r)

	rr.ReadKindAndVerify(rwutil.Kind(requestKindOnLedger))
	rr.ReadN(req.requestID[:])
	rr.ReadN(req.senderAddress[:])
	rr.ReadN(req.targetAddress[:])

	req.requestMetadata, err = RequestMetadataFromBytes(rr.ReadBytes())
	if err != nil {
		return err
	}

	return rr.Err
}

func (req *onLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(requestKindOnLedger))
	ww.WriteN(req.requestID[:])
	ww.WriteN(req.senderAddress[:])
	ww.WriteN(req.targetAddress[:])
	ww.Write(req.requestMetadata)

	return ww.Err
}

func (req *onLedgerRequestData) Allowance() *Assets {
	if req.requestMetadata == nil {
		return NewEmptyAssets()
	}
	return req.requestMetadata.Allowance
}

func (req *onLedgerRequestData) Assets() *Assets {
	return req.assets
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
	outputID := sui.ObjectID{}
	copy(outputID[:], req.requestID[:])

	ret := &onLedgerRequestData{
		requestID:     outputID,
		senderAddress: req.senderAddress.Clone(),
		targetAddress: req.targetAddress.Clone(),
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
	sender := req.SenderAddress()
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

func (req *onLedgerRequestData) SenderAddress() *cryptolib.Address {
	return req.senderAddress
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
	return req.targetAddress
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
