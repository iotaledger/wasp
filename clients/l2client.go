// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type L2Client interface {
	StartNewChain(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		chainOwnerAddress *cryptolib.Address,
		packageID iotago.PackageID,
		stateMetadata []byte,
		initCoinRef *iotago.ObjectRef,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iscmove.AnchorWithRef, error)
	CreateAndSendRequest(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		anchorAddress *iotago.ObjectID,
		assetsBagRef *iotago.ObjectRef,
		msg *iscmove.Message,
		allowance *iscmove.Assets,
		onchainGasBudget uint64,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	ReceiveRequestsAndTransition(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		anchorRef *iotago.ObjectRef,
		reqs []iotago.ObjectRef,
		stateMetadata []byte,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsBagNew(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsBagPlaceCoin(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		assetsBagRef *iotago.ObjectRef,
		coin *iotago.ObjectRef,
		coinType iotajsonrpc.CoinType,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsBagPlaceCoinAmount(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		assetsBagRef *iotago.ObjectRef,
		coin *iotago.ObjectRef,
		coinType iotajsonrpc.CoinType,
		amount uint64,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsDestroyEmpty(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		assetsBagRef *iotago.ObjectRef,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAssetsBagWithBalances(
		ctx context.Context,
		assetsBagID *iotago.ObjectID,
	) (*iscmove.AssetsBagWithBalances, error)
	CreateAndSendRequestWithAssets(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		anchorAddress *iotago.ObjectID,
		assets *iscmove.Assets,
		msg *iscmove.Message,
		allowance *iscmove.Assets,
		onchainGasBudget uint64,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAnchorFromObjectID(
		ctx context.Context,
		anchorObjectID *iotago.ObjectID,
	) (*iscmove.RefWithObject[iscmove.Anchor], error)
	GetRequestFromObjectID(
		ctx context.Context,
		reqID *iotago.ObjectID,
	) (*iscmove.RefWithObject[iscmove.Request], error)
}

var _ L2Client = &iscmoveclient.Client{}
