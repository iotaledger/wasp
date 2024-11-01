package vmtxbuilder

import (
	"slices"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	iscPackage iotago.Address

	// anchorOutput output of the chain
	anchor *isc.StateAnchor

	// already consumed requests, specified by entire Request. It is needed for checking validity
	consumed []isc.OnLedgerRequest

	ptb *iotago.ProgrammableTransactionBuilder

	ownerAddr *cryptolib.Address
}

var _ TransactionBuilder = &AnchorTransactionBuilder{}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(
	iscPackage iotago.Address,
	anchor *isc.StateAnchor,
	ownerAddr *cryptolib.Address,
) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		iscPackage: iscPackage,
		anchor:     anchor,
		ptb:        iotago.NewProgrammableTransactionBuilder(),
		ownerAddr:  ownerAddr,
	}
}

func (txb *AnchorTransactionBuilder) Clone() TransactionBuilder {
	return &AnchorTransactionBuilder{
		anchor:     txb.anchor,
		iscPackage: txb.iscPackage,
		consumed:   slices.Clone(txb.consumed),
		ptb:        txb.ptb.Clone(),
		ownerAddr:  txb.ownerAddr,
	}
}

// ConsumeRequest adds an input to the transaction.
// It panics if transaction cannot hold that many inputs
// All explicitly consumed inputs will hold fixed index in the transaction
// Returns  the amount of baseTokens needed to cover SD costs for the NTs/NFT contained by the request output
func (txb *AnchorTransactionBuilder) ConsumeRequest(req isc.OnLedgerRequest) {
	// TODO we may need to assert the maximum size of the request we can consume here

	txb.consumed = append(txb.consumed, req)
}

func (txb *AnchorTransactionBuilder) SendAssets(target *iotago.Address, assets *isc.Assets) {
	if txb.ptb == nil {
		txb.ptb = iotago.NewProgrammableTransactionBuilder()
	}
	// FIXME allow assets but not only coin balance

	txb.ptb = iscmoveclient.PTBTakeAndTransferCoinBalance(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		target,
		assets,
	)
}

func (txb *AnchorTransactionBuilder) SendCrossChainRequest(targetPackage *iotago.Address, targetAnchor *iotago.Address, assets *isc.Assets, metadata *isc.SendMetadata) {
	if txb.ptb == nil {
		txb.ptb = iotago.NewProgrammableTransactionBuilder()
	}
	txb.ptb = iscmoveclient.PTBAssetsBagNew(txb.ptb, txb.iscPackage, txb.ownerAddr)
	argAssetsBag := txb.ptb.LastCommandResultArg()
	argAnchor := txb.ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()})
	for coinType, coinBalance := range assets.Coins {
		txb.ptb = iscmoveclient.PTBTakeAndPlaceToAssetsBag(txb.ptb, txb.iscPackage, argAnchor, argAssetsBag, coinBalance.Uint64(), coinType.String())
	}
	// TODO set allowance
	allowanceCointypes := txb.ptb.MustForceSeparatePure(&bcs.Option[[]string]{None: true})
	allowanceBalances := txb.ptb.MustForceSeparatePure(&bcs.Option[[]uint64]{None: true})
	txb.ptb = iscmoveclient.PTBCreateAndSendCrossRequest(
		txb.ptb,
		txb.iscPackage,
		*targetAnchor,
		argAssetsBag,
		uint32(metadata.Message.Target.Contract),
		uint32(metadata.Message.Target.EntryPoint),
		metadata.Message.Params,
		allowanceCointypes,
		allowanceBalances,
		metadata.GasBudget,
	)
}

func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateMetadata []byte) iotago.ProgrammableTransaction {
	if txb.ptb == nil {
		txb.ptb = iotago.NewProgrammableTransactionBuilder()
	}
	// we have to discard the current txb to avoid reusing an ObjectRef
	defer func() { txb.ptb = nil }()
	ptb := iscmoveclient.PTBReceiveRequestAndTransition(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		lo.Map(txb.consumed, func(r isc.OnLedgerRequest, _ int) iotago.ObjectRef { return r.RequestRef() }),
		lo.Map(txb.consumed, func(r isc.OnLedgerRequest, _ int) *iscmove.AssetsBagWithBalances { return r.AssetsBag() }),
		stateMetadata,
	)
	return ptb.Finish()
}

func NewRotationTransaction(anchorRef *iotago.ObjectRef, rotationAddress *iotago.Address) (*iotago.ProgrammableTransaction, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.TransferObject(rotationAddress, anchorRef); err != nil {
		return nil, err
	}
	pt := ptb.Finish()
	return &pt, nil
}
