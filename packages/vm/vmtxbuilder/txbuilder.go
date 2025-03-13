package vmtxbuilder

import (
	"fmt"
	"slices"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
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

	rotateToAddr *iotago.Address
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
	txb.ptb = iscmoveclient.PTBTakeAndTransferCoinBalance(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		target,
		assets.AsISCMove(),
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

func (txb *AnchorTransactionBuilder) RotationTransaction(rotationAddress *iotago.Address) {
	txb.rotateToAddr = rotationAddress
}

func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateMetadata []byte, topUpAmount uint64) iotago.ProgrammableTransaction {
	if txb.ptb == nil {
		txb.ptb = iotago.NewProgrammableTransactionBuilder()
	}
	ptb := iscmoveclient.PTBReceiveRequestsAndTransition(
		txb.ptb,
		txb.iscPackage,
		txb.ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
		lo.Map(txb.consumed, func(r isc.OnLedgerRequest, _ int) iotago.ObjectRef { return r.RequestRef() }),
		lo.Map(txb.consumed, func(r isc.OnLedgerRequest, _ int) *iscmove.AssetsBagWithBalances { return r.AssetsBag() }),
		stateMetadata,
		topUpAmount,
	)

	if txb.rotateToAddr != nil {
		ptb.Command(iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: txb.anchor.GetObjectRef()}),
				},
				Address: ptb.MustForceSeparatePure(txb.rotateToAddr),
			},
		})
		ptb.Command(iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{
					iotago.GetArgumentGasCoin(),
				},
				Address: ptb.MustForceSeparatePure(txb.rotateToAddr),
			},
		})
	}
	return ptb.Finish()
}

func (txb *AnchorTransactionBuilder) ViewPTB() *iotago.ProgrammableTransactionBuilder {
	return txb.ptb.Clone()
}

func (txb *AnchorTransactionBuilder) CheckTransactionSize() error {
	const maxTxSizeBytes = 128 * 1024
	const maxInputObjects = 2048
	const maxProgrammableTxCommands = 1024
	ptb := txb.ViewPTB()

	emptyStateMetadata := make([]byte, 32)
	tx := txb.BuildTransactionEssence(emptyStateMetadata, 10)

	if len(tx.Inputs) > maxInputObjects {
		return fmt.Errorf("tx input len: %d, exceed max_input_objects: %w", len(tx.Inputs), vmexceptions.ErrMaxTransactionSizeExceeded)
	}
	if len(ptb.Commands) > maxProgrammableTxCommands {
		return fmt.Errorf("tx command len: %d, exceed max_programmable_tx_commands: %w", len(tx.Commands), vmexceptions.ErrMaxTransactionSizeExceeded)
	}

	b, _ := bcs.Marshal(&tx)
	if len(b) > maxTxSizeBytes {
		return fmt.Errorf("ptb serialized size: %d, exceed max_tx_size_bytes: %w", maxTxSizeBytes, vmexceptions.ErrMaxTransactionSizeExceeded)
	}
	return nil
}
