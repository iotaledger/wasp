package migrations

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/core/types"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/samber/lo"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func migrateOnLedgerContractIdentity(request old_isc.OnLedgerRequest) isc.ContractIdentity {
	var oldContractIdentity old_isc.ContractIdentity
	newContractIdentity := isc.EmptyContractIdentity()

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

func migrateOnLedgerRequest(request old_isc.OnLedgerRequest, oldChainID old_isc.ChainID) isc.Request {
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
		Message:        migrateContractCall(oldChainID, request.CallTarget().Contract, request.CallTarget().EntryPoint, request.Params()),
		AllowanceBCS:   bcs.MustMarshal(OldAssetsToNewAssets(request.Allowance())),
		GasBudget:      gasBudget,
	}

	return isc.NewOnLedgerRequestData(requestRef, senderAddress, targetAddress, assets, &iscmove.AssetsBagWithBalances{
		AssetsBag: *fakeAssetsBag,
		Assets:    *iscmove.NewEmptyAssets(),
	}, requestMetadata)
}

func migrateOffLedgerRequest(req old_isc.OffLedgerRequest, oldChainID old_isc.ChainID) isc.Request {
	if evmCallMsg := req.EVMCallMsg(); evmCallMsg != nil {
		// read unexported tx field of evmOffLedgerTxRequest struct
		var tx *types.Transaction
		{
			reqTxField := reflect.ValueOf(req).Elem().Field(1)
			reqTxField = reflect.NewAt(reqTxField.Type(), unsafe.Pointer(reqTxField.UnsafeAddr())).Elem()
			txValue := reflect.ValueOf(&tx).Elem()
			txValue.Set(reqTxField)
		}

		return lo.Must(isc.NewEVMOffLedgerTxRequest(isc.ChainID{}, tx))
	} else {
		message := migrateContractCall(oldChainID, req.CallTarget().Contract, req.CallTarget().EntryPoint, req.Params())
		nonce := req.Nonce()
		gasbudget, _ := req.GasBudget()
		allowance := OldAssetsToNewAssets(req.Allowance())

		// read unexported signature of OffLedgerRequestData struct
		var sig struct {
			publicKey *cryptolib.PublicKey
			signature []byte
		}
		{
			reqSigField := reflect.ValueOf(req).Elem().Field(8)
			reqSigField = reflect.NewAt(reqSigField.Type(), unsafe.Pointer(reqSigField.UnsafeAddr())).Elem()
			sigValue := reflect.ValueOf(&sig).Elem()
			sigValue.Set(reqSigField)
		}
		newRequest := isc.NewOffLedgerRequestsRaw(
			allowance,
			isc.ChainID{},
			message,
			gasbudget,
			nonce,
			cryptolib.NewSignature(sig.publicKey, sig.signature),
		)
		return &newRequest
	}
}

func MigrateSingleRequest(req old_isc.Request, oldChainID old_isc.ChainID) isc.Request {
	switch req.(type) {
	case old_isc.OnLedgerRequest:
		return migrateOnLedgerRequest(req.(old_isc.OnLedgerRequest), oldChainID)

	case old_isc.OffLedgerRequest:
		return migrateOffLedgerRequest(req.(old_isc.OffLedgerRequest), oldChainID)

	case old_isc.UnsignedOffLedgerRequest:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))

	case old_isc.ImpersonatedOffLedgerRequest:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))

	default:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))
	}

	return nil
}
