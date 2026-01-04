package client

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

// IotaClient is an interface that abstracts the common IOTA client operations.
// Implemented by the JSON-RPC and GraphQL clients.
type IotaClient interface {
	// Read API
	GetDynamicFieldObject(
		ctx context.Context,
		req GetDynamicFieldObjectRequest,
	) (*iotajsonrpc.IotaObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req GetDynamicFieldsRequest,
	) (*iotajsonrpc.DynamicFieldPage, error)
	GetOwnedObjects(
		ctx context.Context,
		req GetOwnedObjectsRequest,
	) (*iotajsonrpc.ObjectsPage, error)
	QueryTransactionBlocks(
		ctx context.Context,
		req QueryTransactionBlocksRequest,
	) (*iotajsonrpc.TransactionBlocksPage, error)
	DevInspectTransactionBlock(
		ctx context.Context,
		req DevInspectTransactionBlockRequest,
	) (*iotajsonrpc.DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		req DryRunTransactionRequest,
	) (*iotajsonrpc.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req ExecuteTransactionBlockRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error)

	// Transaction Builder API
	MergeCoins(
		ctx context.Context,
		req MergeCoinsRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	MoveCall(
		ctx context.Context,
		req MoveCallRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	PayAllIota(
		ctx context.Context,
		req PayAllIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	PayIota(
		ctx context.Context,
		req PayIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	Publish(
		ctx context.Context,
		req PublishRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	SplitCoin(
		ctx context.Context,
		req SplitCoinRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req TransferObjectRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	TransferIota(
		ctx context.Context,
		req TransferIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)

	// Coin Query API
	GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error)
	GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetBalance(ctx context.Context, req GetBalanceRequest) (*iotajsonrpc.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error)
	GetCoins(ctx context.Context, req GetCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error)

	// Extended API
	GetObject(ctx context.Context, req GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error)
	GetTransactionBlock(ctx context.Context, req GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req TryGetPastObjectRequest,
	) (*iotajsonrpc.IotaPastObjectResponse, error)

	// Utility methods
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *iotago.Address,
		targetAmount uint64,
		gasAmount uint64,
	) (iotajsonrpc.Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		req *SignAndExecuteTransactionRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	UpdateObjectRef(
		ctx context.Context,
		ref *iotago.ObjectRef,
	) (*iotago.ObjectRef, error)
	MintToken(
		ctx context.Context,
		signer iotasigner.Signer,
		packageID *iotago.PackageID,
		tokenName string,
		treasuryCap *iotago.ObjectRef,
		mintAmount uint64,
		options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	SignAndExecuteTxWithRetry(
		ctx context.Context,
		signer iotasigner.Signer,
		pt iotago.ProgrammableTransaction,
		gasCoin *iotago.ObjectRef,
		gasBudget uint64,
		gasPrice uint64,
		options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
}
