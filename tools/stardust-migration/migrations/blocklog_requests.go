package migrations

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func migrateOnLedgerContractIdentity(request old_isc.OnLedgerRequest) isc.ContractIdentity {
	var oldContractIdentity old_isc.ContractIdentity
	var newContractIdentity isc.ContractIdentity = isc.EmptyContractIdentity()

	var onLedgerRequestData old_isc.OnLedgerRequestData

	if o, ok := request.(*old_isc.OnLedgerRequestData); ok {
		onLedgerRequestData = *o
	}

	if o, ok := request.(*old_isc.RetryOnLedgerRequest); ok {
		if data, ok := o.OnLedgerRequest.(*old_isc.OnLedgerRequestData); ok {
			onLedgerRequestData = *data
		} else {
			panic("Failed to cast RetryOnLedger to OnLedgerRequestData")
		}
	}

	if onLedgerRequestData.RequestMetadataRaw() != nil && !onLedgerRequestData.RequestMetadataRaw().SenderContract.Empty() {
		oldContractIdentity = onLedgerRequestData.RequestMetadataRaw().SenderContract
		newContractIdentity = isc.NewContractIdentity(byte(oldContractIdentity.Kind), oldContractIdentity.EvmAddr, OldHnameToNewHname(oldContractIdentity.HnameRaw()))
	}

	return newContractIdentity
}

func migrateOnLedgerRequest(request old_isc.OnLedgerRequest /*, oldChainID old_isc.ChainID, newChainID isc.ChainID*/) isc.Request {
	requestRef := iotago.ObjectRef{
		ObjectID: (*iotago.ObjectID)(request.ID().Bytes()),
		Version:  0,
		Digest:   lo.Must(iotago.NewDigest("MIGRATED_FROM_STARDUST")),
	}
	gasBudget, _ := request.GasBudget()

	var senderAddress *cryptolib.Address
	if request.Output().FeatureSet().SenderFeature() != nil {
		senderAddress = OldIotaGoAddressToCryptoLibAddress(request.Output().FeatureSet().SenderFeature().Address)
	}

	targetAddress := OldIotaGoAddressToCryptoLibAddress(request.TargetAddress())
	assets := OldAssetsToNewAssets(request.Assets())
	// For now selecting the transactionID of the request to fake the AssetBag ref.
	// I would prefer the OutputID as it would hold the assets, but it has a length of 34bytes, the ObjectID has 32.
	fakeAssetsBag := &iscmove.AssetsBag{Size: 0, ID: *iotago.MustObjectIDFromHex(request.OutputID().TransactionID().ToHex())}

	requestMetadata := &isc.RequestMetadata{
		SenderContract: migrateOnLedgerContractIdentity(request),
		Message: isc.Message{
			Target: isc.CallTarget{
				Contract:   OldHnameToNewHname(request.CallTarget().Contract),
				EntryPoint: OldHnameToNewHname(request.CallTarget().EntryPoint),
			},
			//Params: request.Params(),
		},
		Allowance: OldAssetsToNewAssets(request.Allowance()),
		GasBudget: gasBudget,
	}

	return isc.NewOnLedgerRequestData(requestRef, senderAddress, targetAddress, assets, fakeAssetsBag, requestMetadata)
}

func migrateOffLedgerRequest(req old_isc.OffLedgerRequest, oldChainID old_isc.ChainID, newChainID isc.ChainID) isc.Request {
	message := isc.Message{
		Target: isc.CallTarget{
			Contract:   OldHnameToNewHname(req.CallTarget().Contract),
			EntryPoint: OldHnameToNewHname(req.CallTarget().EntryPoint),
		},
		//Params: request.Params(),
	}
	nonce := req.Nonce()
	gasbudget, _ := req.GasBudget()
	allowance := OldAssetsToNewAssets(req.Allowance())

	// TODO: Implement old_isc.Request -> Signature()
	newRequest := isc.NewOffLedgerRequestsRaw(allowance, newChainID, message, gasbudget, nonce, &cryptolib.Signature{})

	return &newRequest
}

func MigrateSingleRequest(req old_isc.Request, oldChainID old_isc.ChainID, newChainID isc.ChainID) isc.Request {
	switch req.(type) {
	case old_isc.OnLedgerRequest:
		return migrateOnLedgerRequest(req.(old_isc.OnLedgerRequest))

	case old_isc.OffLedgerRequest:
		return migrateOffLedgerRequest(req.(old_isc.OffLedgerRequest), oldChainID, newChainID)

	case old_isc.UnsignedOffLedgerRequest:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))

	case old_isc.ImpersonatedOffLedgerRequest:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))

	default:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))
	}

	return nil
}
