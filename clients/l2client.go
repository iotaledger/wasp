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
		packageID iotago.PackageID,
		stateMetadata []byte,
		initCoinRef *iotago.ObjectRef,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*iscmove.AnchorWithRef, error)
	CreateAndSendRequest(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		anchorAddress *iotago.ObjectID,
		assetsBagRef *iotago.ObjectRef,
		iscContractName uint32,
		iscFunctionName uint32,
		args [][]byte,
		allowanceArray []iscmove.CoinAllowance,
		onchainGasBudget uint64,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	ReceiveRequestAndTransition(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		anchorRef *iotago.ObjectRef,
		reqs []iotago.ObjectRef,
		stateMetadata []byte,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsBagNew(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
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
		devMode bool,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	AssetsDestroyEmpty(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID iotago.PackageID,
		assetsBagRef *iotago.ObjectRef,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAssetsBagWithBalances(
		ctx context.Context,
		assetsBagID *iotago.ObjectID,
	) (*iscmove.AssetsBagWithBalances, error)
}

var _ L2Client = &iscmoveclient.Client{}
