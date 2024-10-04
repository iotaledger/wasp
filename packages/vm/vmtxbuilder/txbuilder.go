package vmtxbuilder

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	iscPackage sui.Address

	// anchorOutput output of the chain
	anchor *isc.StateAnchor

	// already consumed requests, specified by entire Request. It is needed for checking validity
	consumed []isc.OnLedgerRequest

	ptb *sui.ProgrammableTransactionBuilder

	ownerAddr *cryptolib.Address
}

var _ TransactionBuilder = &AnchorTransactionBuilder{}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(
	iscPackage sui.Address,
	anchor *isc.StateAnchor,
	ownerAddr *cryptolib.Address,
) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		iscPackage: iscPackage,
		anchor:     anchor,
		ptb:        sui.NewProgrammableTransactionBuilder(),
		ownerAddr:  ownerAddr,
	}
}

func (txb *AnchorTransactionBuilder) Clone() TransactionBuilder {
	a := *txb.anchor
	newConsumed := make([]isc.OnLedgerRequest, len(txb.consumed))
	for i, v := range txb.consumed {
		newConsumed[i] = v.Clone()
	}
	return &AnchorTransactionBuilder{
		anchor:   &a,
		consumed: newConsumed,
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

func (txb *AnchorTransactionBuilder) SendAssets(target *sui.Address, assets *isc.Assets) {
	if txb.ptb == nil {
		txb.ptb = sui.NewProgrammableTransactionBuilder()
	}
	// FIXME allow assets but not only coin balance

	txb.ptb = iscmoveclient.PTBTakeAndTransferCoinBalance(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		target,
		assets,
	)
}

func (txb *AnchorTransactionBuilder) SendCrossChainRequest(targetPackage *sui.Address, targetAnchor *sui.Address, assets *isc.Assets, metadata *isc.SendMetadata) {
	if txb.ptb == nil {
		txb.ptb = sui.NewProgrammableTransactionBuilder()
	}
	txb.ptb = iscmoveclient.PTBAssetsBagNew(txb.ptb, txb.iscPackage, txb.ownerAddr)
	argAssetsBag := txb.ptb.LastCommandResultArg()
	argAnchor := txb.ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()})
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

func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateMetadata []byte) sui.ProgrammableTransaction {
	if txb.ptb == nil {
		txb.ptb = sui.NewProgrammableTransactionBuilder()
	}
	// we have to discard the current txb to avoid reusing an ObjectRef
	defer func() { txb.ptb = nil }()
	ptb := iscmoveclient.PTBReceiveRequestAndTransition(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		onRequestsToRequestRefs(txb.consumed),
		onRequestsToAssetsBagMap(txb.consumed),
		stateMetadata,
	)
	return ptb.Finish()
}

func onRequestsToRequestRefs(reqs []isc.OnLedgerRequest) []sui.ObjectRef {
	refs := make([]sui.ObjectRef, len(reqs))
	for i, req := range reqs {
		refs[i] = req.RequestRef()
	}
	return refs
}

func onRequestsToAssetsBagMap(reqs []isc.OnLedgerRequest) map[sui.ObjectRef]*iscmove.AssetsBagWithBalances {
	m := make(map[sui.ObjectRef]*iscmove.AssetsBagWithBalances)
	for _, req := range reqs {
		assetsBagWithBalances := &iscmove.AssetsBagWithBalances{
			AssetsBag: *req.AssetsBag(),
			Balances:  make(iscmove.AssetsBagBalances),
		}
		assets := req.Assets()
		for k, v := range assets.Coins {
			assetsBagWithBalances.Balances[suijsonrpc.CoinType(k)] = &suijsonrpc.Balance{
				CoinType:     suijsonrpc.CoinType(k),
				TotalBalance: uint64(v),
			}
		}
		m[req.RequestRef()] = assetsBagWithBalances

	}
	return m
}

func NewRotationTransaction(rotationAddress *sui.Address) (*sui.TransactionData, error) {
	panic("txbuilder.NewRotationTransaction -- implement") // TODO: Implement.
}
