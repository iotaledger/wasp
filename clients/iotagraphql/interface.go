package iotagraphql

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

type IotaClient interface {
	// Read API
	GetDynamicFieldObject(
		ctx context.Context,
		req GetDynamicFieldObjectRequest,
	) (*IotaObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req GetDynamicFieldsRequest,
	) (*DynamicFieldPage, error)
	GetOwnedObjects(
		ctx context.Context,
		req GetOwnedObjectsRequest,
	) (*ObjectsPage, error)
	QueryTransactionBlocks(
		ctx context.Context,
		req QueryTransactionBlocksRequest,
	) (*TransactionBlocksPage, error)
	DevInspectTransactionBlock(
		ctx context.Context,
		req DevInspectTransactionBlockRequest,
	) (*DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		req DryRunTransactionRequest,
	) (*DryRunResult, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req ExecuteTransactionBlockRequest,
	) (*IotaTransactionBlockResponse, error)
	GetLatestIotaSystemState(ctx context.Context) (*IotaSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*BigInt, error)

	// Transaction Builder API
	MergeCoins(
		ctx context.Context,
		req MergeCoinsRequest,
	) (*TransactionBytes, error)
	MoveCall(
		ctx context.Context,
		req MoveCallRequest,
	) (*TransactionBytes, error)
	PayAllIota(
		ctx context.Context,
		req PayAllIotaRequest,
	) (*TransactionBytes, error)
	PayIota(
		ctx context.Context,
		req PayIotaRequest,
	) (*TransactionBytes, error)
	Publish(
		ctx context.Context,
		req PublishRequest,
	) (*TransactionBytes, error)
	SplitCoin(
		ctx context.Context,
		req SplitCoinRequest,
	) (*TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req TransferObjectRequest,
	) (*TransactionBytes, error)
	TransferIota(
		ctx context.Context,
		req TransferIotaRequest,
	) (*TransactionBytes, error)

	// Coin Query API
	GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*Balance, error)
	GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*CoinPage, error)
	GetBalance(ctx context.Context, req GetBalanceRequest) (*Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*IotaCoinMetadata, error)
	GetCoins(ctx context.Context, req GetCoinsRequest) (*CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType string) (*Supply, error)

	// Extended API
	GetObject(ctx context.Context, req GetObjectRequest) (*IotaObjectResponse, error)
	GetTransactionBlock(ctx context.Context, req GetTransactionBlockRequest) (*IotaTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req TryGetPastObjectRequest,
	) (*IotaPastObjectResponse, error)

	// Utility methods
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *iotago.Address,
		targetAmount uint64,
		gasAmount uint64,
	) (Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		req *SignAndExecuteTransactionRequest,
	) (*IotaTransactionBlockResponse, error)
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
		options *IotaTransactionBlockResponseOptions,
	) (*IotaTransactionBlockResponse, error)
	SignAndExecuteTxWithRetry(
		ctx context.Context,
		signer iotasigner.Signer,
		pt iotago.ProgrammableTransaction,
		gasCoin *iotago.ObjectRef,
		gasBudget uint64,
		gasPrice uint64,
		options *IotaTransactionBlockResponseOptions,
	) (*IotaTransactionBlockResponse, error)
}
