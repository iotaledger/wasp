// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iota-go/sui"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type L2Client interface {
	StartNewChain(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		stateMetadata []byte,
		initCoinRef *sui.ObjectRef,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*iscmove.AnchorWithRef, error)
	CreateAndSendRequest(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		anchorAddress *sui.ObjectID,
		assetsBagRef *sui.ObjectRef,
		iscContractName uint32,
		iscFunctionName uint32,
		args [][]byte,
		allowanceArray []iscmove.CoinAllowance,
		onchainGasBudget uint64,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	ReceiveRequestAndTransition(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		anchorRef *sui.ObjectRef,
		reqs []sui.ObjectRef,
		stateMetadata []byte,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	AssetsBagNew(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	AssetsBagPlaceCoin(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		assetsBagRef *sui.ObjectRef,
		coin *sui.ObjectRef,
		coinType suijsonrpc2.CoinType,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	AssetsDestroyEmpty(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID sui.PackageID,
		assetsBagRef *sui.ObjectRef,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	GetAssetsBagWithBalances(
		ctx context.Context,
		assetsBagID *sui.ObjectID,
	) (*iscmove.AssetsBagWithBalances, error)
}

var _ L2Client = &iscmoveclient.Client{}
