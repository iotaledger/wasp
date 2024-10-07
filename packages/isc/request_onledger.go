package isc

import (
	"fmt"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type onLedgerRequestData struct {
	requestRef      sui.ObjectRef      `bcs:""`
	senderAddress   *cryptolib.Address `bcs:""`
	targetAddress   *cryptolib.Address `bcs:""`
	assets          *Assets
	assetsBag       *iscmove.AssetsBag
	requestMetadata *RequestMetadata `bcs:""`
}

var (
	_ Request         = new(onLedgerRequestData)
	_ OnLedgerRequest = new(onLedgerRequestData)
	_ Calldata        = new(onLedgerRequestData)
)

func OnLedgerFromRequest(request *iscmove.RefWithObject[iscmove.Request], anchorAddress *cryptolib.Address) (OnLedgerRequest, error) {
	r := &onLedgerRequestData{
		requestRef:    request.ObjectRef,
		senderAddress: request.Object.Sender,
		targetAddress: anchorAddress,
		assetsBag:     &request.Object.AssetsBag.AssetsBag,
		requestMetadata: &RequestMetadata{
			SenderContract: ContractIdentity{},
			Message: Message{
				Target: CallTarget{
					Contract:   Hname(request.Object.Message.Contract),
					EntryPoint: Hname(request.Object.Message.Function),
				},
				Params: CallArguments{},
			},
			Allowance: NewEmptyAssets(),
			GasBudget: 0,
		},
		assets: AssetsFromAssetsBagWithBalances(request.Object.AssetsBag),
	}

	return r, nil
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
	return bcs.MustMarshal(req)
}

func (req *onLedgerRequestData) Message() Message {
	if req.requestMetadata == nil {
		return Message{}
	}
	return req.requestMetadata.Message
}

func (req *onLedgerRequestData) Clone() OnLedgerRequest {
	outputRef := sui.ObjectRefFromBytes(req.requestRef.Bytes())

	ret := &onLedgerRequestData{
		requestRef:    *outputRef,
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
	return RequestID(*req.requestRef.ObjectID)
}

func (req *onLedgerRequestData) IsOffLedger() bool {
	return false
}

func (req *onLedgerRequestData) RequestID() sui.ObjectID {
	return *req.requestRef.ObjectID
}

func (req *onLedgerRequestData) SenderAccount() AgentID {
	sender := req.SenderAddress()
	if sender == nil {
		return nil
	}
	if req.requestMetadata != nil && !req.requestMetadata.SenderContract.Empty() {
		chainID := ChainIDFromAddress(sender)
		return req.requestMetadata.SenderContract.AgentID(chainID)
	}
	return NewAddressAgentID(sender)
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

func (req *onLedgerRequestData) RequestRef() sui.ObjectRef {
	return req.requestRef
}

func (req *onLedgerRequestData) AssetsBag() *iscmove.AssetsBag {
	return req.assetsBag
}

func (req *onLedgerRequestData) TargetAddress() *cryptolib.Address {
	return req.targetAddress
}

func (req *onLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}
