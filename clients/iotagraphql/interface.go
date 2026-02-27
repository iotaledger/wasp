package iotagraphql

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

type IotaClient interface {
	GetDynamicFieldObject(
		ctx context.Context,
		req GetDynamicFieldObjectRequest,
	) (*GetDynamicFieldObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req GetDynamicFieldsRequest,
	) (*graphqltypes.GetDynamicFieldsResponse, error)
	GetOwnedObjects(
		ctx context.Context,
		req GetOwnedObjectsRequest,
	) (*graphqltypes.GetOwnedObjectsResponse, error)
	DryRunTransaction(
		ctx context.Context,
		txDataBytes iotago.Base64Data,
	) (*graphqltypes.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		txDataBytes iotago.Base64Data,
		signatures []*iotasigner.Signature,
	) (*graphqltypes.ExecuteTransactionBlockResponse, error)
	GetLatestIotaSystemState(ctx context.Context) (*GetLatestIotaSystemStateResponse, error)
	GetReferenceGasPrice(ctx context.Context) (*BigInt, error)

	MergeCoins(
		ctx context.Context,
		req MergeCoinsRequest,
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
	TransferIota(
		ctx context.Context,
		req TransferIotaRequest,
	) (*TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req TransferObjectRequest,
	) (*TransactionBytes, error)

	GetAllBalances(ctx context.Context, owner iotago.Address) ([]*Balance, error)
	GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*GetAllCoinsResponse, error)
	GetBalance(ctx context.Context, req GetBalanceRequest) (*Balance, error)
	GetCoinMetadata(ctx context.Context, coinType CoinType) (*IotaCoinMetadata, error)
	GetCoins(ctx context.Context, req GetCoinsRequest) (*GetCoinsResponse, error)
	GetTotalSupply(ctx context.Context, coinType CoinType) (*Supply, error)
	GetObject(ctx context.Context, objectID iotago.ObjectID) (*graphqltypes.GetObjectResponse, error)
	GetTransactionBlock(ctx context.Context, digest iotago.TransactionDigest) (*graphqltypes.GetTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		objectID iotago.ObjectID,
		version uint64,
	) (*TryGetPastObjectResponse, error)
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address iotago.Address,
		targetAmount uint64,
		gasAmount uint64,
	) (Coins, error)

	SignAndExecuteTransaction(
		ctx context.Context,
		txnBytes []byte,
		signer iotasigner.Signer,
	) (*graphqltypes.ExecuteTransactionBlockResponse, error)
	UpdateObjectRef(
		ctx context.Context,
		ref *iotago.ObjectRef,
	) (*iotago.ObjectRef, error)
	MintToken(
		ctx context.Context,
		signer iotasigner.Signer,
		packageID iotago.PackageID,
		tokenName string,
		treasuryCap *iotago.ObjectRef,
		mintAmount uint64,
		maxRetries int,
	) (*graphqltypes.ExecuteTransactionBlockResponse, error)
	SignAndExecuteTxWithRetry(
		ctx context.Context,
		signer iotasigner.Signer,
		pt iotago.ProgrammableTransaction,
		gasCoin *iotago.ObjectRef,
		gasBudget uint64,
		gasPrice uint64,
	) (*ExecuteTransactionBlockResponse, error)
	RequestFundsFromFaucet(ctx context.Context, address iotago.Address) error
}

var _ IotaClient = (*GraphQLClient)(nil)
