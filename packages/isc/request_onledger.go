package isc

import (
	"fmt"

	"github.com/ethereum/go-ethereum"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type OnLedgerRequestData struct {
	requestRef      iotago.ObjectRef   `bcs:"export"`
	senderAddress   *cryptolib.Address `bcs:"export"`
	targetAddress   *cryptolib.Address `bcs:"export"`
	assets          *Assets            `bcs:"export"`
	assetsBag       *iscmove.AssetsBag `bcs:"export"`
	requestMetadata *RequestMetadata   `bcs:"export"`
}

var (
	_ Request         = new(OnLedgerRequestData)
	_ OnLedgerRequest = new(OnLedgerRequestData)
	_ Calldata        = new(OnLedgerRequestData)
)

func OnLedgerFromRequest(request *iscmove.RefWithObject[iscmove.Request], anchorAddress *cryptolib.Address) (OnLedgerRequest, error) {
	assets, err := AssetsFromAssetsBagWithBalances(&request.Object.AssetsBag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assets from AssetsBag: %w", err)
	}
	allowance, err := AssetsFromISCMove(&request.Object.Allowance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowance: %w", err)
	}
	r := &OnLedgerRequestData{
		requestRef:    request.ObjectRef,
		senderAddress: request.Object.Sender,
		targetAddress: anchorAddress,
		assets:        assets,
		assetsBag:     &request.Object.AssetsBag.AssetsBag,
		requestMetadata: &RequestMetadata{
			SenderContract: ContractIdentity{},
			Message: Message{
				Target: CallTarget{
					Contract:   Hname(request.Object.Message.Contract),
					EntryPoint: Hname(request.Object.Message.Function),
				},
				Params: request.Object.Message.Args,
			},
			Allowance: allowance,
			GasBudget: request.Object.GasBudget,
		},
	}
	return r, nil
}

func (req *OnLedgerRequestData) Allowance() *Assets {
	if req.requestMetadata == nil {
		return NewEmptyAssets()
	}
	return req.requestMetadata.Allowance
}

func (req *OnLedgerRequestData) Assets() *Assets {
	return req.assets
}

func (req *OnLedgerRequestData) Bytes() []byte {
	var r Request = req
	return bcs.MustMarshal(&r)
}

func (req *OnLedgerRequestData) Message() Message {
	if req.requestMetadata == nil {
		return Message{}
	}
	return req.requestMetadata.Message
}

func (req *OnLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	if req.requestMetadata == nil {
		return 0, false
	}
	return req.requestMetadata.GasBudget, false
}

func (req *OnLedgerRequestData) ID() RequestID {
	return RequestID(*req.requestRef.ObjectID)
}

func (req *OnLedgerRequestData) IsOffLedger() bool {
	return false
}

func (req *OnLedgerRequestData) RequestID() iotago.ObjectID {
	return *req.requestRef.ObjectID
}

func (req *OnLedgerRequestData) SenderAccount() AgentID {
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

func (req *OnLedgerRequestData) SenderAddress() *cryptolib.Address {
	return req.senderAddress
}

func (req *OnLedgerRequestData) String() string {
	metadata := req.requestMetadata
	if metadata == nil {
		return "onledger request without metadata"
	}
	return fmt.Sprintf("OnLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, Assets: %v, GasBudget: %d }",
		req.ID().String(),
		req.SenderAddress().String(),
		metadata.Message.Target.Contract.String(),
		metadata.Message.Target.EntryPoint.String(),
		metadata.Message.Params.String(),
		req.assets,
		metadata.GasBudget,
	)
}

func (req *OnLedgerRequestData) RequestRef() iotago.ObjectRef {
	return req.requestRef
}

func (req *OnLedgerRequestData) AssetsBag() *iscmove.AssetsBagWithBalances {
	return req.assets.AsAssetsBagWithBalances(req.assetsBag)
}

func (req *OnLedgerRequestData) TargetAddress() *cryptolib.Address {
	return req.targetAddress
}

func (req *OnLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}
