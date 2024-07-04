// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type L2Client interface {
	StartNewChain(
		ctx context.Context,
		signer cryptolib.Signer,
		packageID *sui.PackageID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
		treasuryCap *suijsonrpc.SuiObjectResponse,
	) (*iscmove.Anchor, error)
	SendCoin(
		ctx context.Context,
		signer cryptolib.Signer,
		anchorPackageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		coinType string,
		coinObject *sui.ObjectID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	ReceiveCoin(
		ctx context.Context,
		signer cryptolib.Signer,
		anchorPackageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		coinType string,
		receivingCoinObject *sui.ObjectID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	GetAssets(
		ctx context.Context,
		anchorPackageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
	) (*iscmove.Assets, error)
	CreateRequest(
		ctx context.Context,
		signer cryptolib.Signer,
		packageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		iscContractName string,
		iscFunctionName string,
		args [][]byte,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	SendRequest(
		ctx context.Context,
		signer cryptolib.Signer,
		packageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		reqObjID *sui.ObjectID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	ReceiveRequest(
		ctx context.Context,
		signer cryptolib.Signer,
		packageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		reqObjID *sui.ObjectID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
}

var _ L2Client = &iscmove.Client{}

func NewL2Client(config iscmove.Config) L2Client {
	return iscmove.NewClient(config)
}
