package client

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

// IotaClient is an interface that abstracts the common IOTA client operations.
// Both iotaclient.Client and iotagraphql.GraphQLClient implement this interface.
// FIXME this is a temporary interface to facilitate the migration from iotaclient to iotagraphql.
type IotaClient interface {
	// Read API
	GetDynamicFieldObject(
		ctx context.Context,
		req iotaclient.GetDynamicFieldObjectRequest,
	) (*iotajsonrpc.IotaObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req iotaclient.GetDynamicFieldsRequest,
	) (*iotajsonrpc.DynamicFieldPage, error)
	GetOwnedObjects(
		ctx context.Context,
		req iotaclient.GetOwnedObjectsRequest,
	) (*iotajsonrpc.ObjectsPage, error)
	QueryEvents(
		ctx context.Context,
		req iotaclient.QueryEventsRequest,
	) (*iotajsonrpc.EventPage, error)
	QueryTransactionBlocks(
		ctx context.Context,
		req iotaclient.QueryTransactionBlocksRequest,
	) (*iotajsonrpc.TransactionBlocksPage, error)
	ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error)
	ResolveNameServiceNames(
		ctx context.Context,
		req iotaclient.ResolveNameServiceNamesRequest,
	) (*iotajsonrpc.IotaNamePage, error)
	DevInspectTransactionBlock(
		ctx context.Context,
		req iotaclient.DevInspectTransactionBlockRequest,
	) (*iotajsonrpc.DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		req iotaclient.DryRunTransactionRequest,
	) (*iotajsonrpc.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req iotaclient.ExecuteTransactionBlockRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetCommitteeInfo(
		ctx context.Context,
		epoch *iotajsonrpc.BigInt,
	) (*iotajsonrpc.CommitteeInfo, error)
	GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error)
	GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error)
	GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error)
	GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error)

	// Transaction Builder API
	BatchTransaction(
		ctx context.Context,
		req iotaclient.BatchTransactionRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	MergeCoins(
		ctx context.Context,
		req iotaclient.MergeCoinsRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	MoveCall(
		ctx context.Context,
		req iotaclient.MoveCallRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	Pay(
		ctx context.Context,
		req iotaclient.PayRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	PayAllIota(
		ctx context.Context,
		req iotaclient.PayAllIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	PayIota(
		ctx context.Context,
		req iotaclient.PayIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	Publish(
		ctx context.Context,
		req iotaclient.PublishRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	RequestAddStake(
		ctx context.Context,
		req iotaclient.RequestAddStakeRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	RequestWithdrawStake(
		ctx context.Context,
		req iotaclient.RequestWithdrawStakeRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	SplitCoin(
		ctx context.Context,
		req iotaclient.SplitCoinRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	SplitCoinEqual(
		ctx context.Context,
		req iotaclient.SplitCoinEqualRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req iotaclient.TransferObjectRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	TransferIota(
		ctx context.Context,
		req iotaclient.TransferIotaRequest,
	) (*iotajsonrpc.TransactionBytes, error)

	// Coin Query API
	GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error)
	GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error)
	GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error)

	// Extended API
	GetChainIdentifier(ctx context.Context) (string, error)
	GetCheckpoint(ctx context.Context, checkpointID *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error)
	GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error)
	GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error)
	GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error)
	GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error)
	GetProtocolConfig(
		ctx context.Context,
		version *iotajsonrpc.BigInt,
	) (*iotajsonrpc.ProtocolConfig, error)
	GetTotalTransactionBlocks(ctx context.Context) (string, error)
	GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.IotaObjectResponse, error)
	MultiGetTransactionBlocks(
		ctx context.Context,
		req iotaclient.MultiGetTransactionBlocksRequest,
	) ([]*iotajsonrpc.IotaTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req iotaclient.TryGetPastObjectRequest,
	) (*iotajsonrpc.IotaPastObjectResponse, error)
	TryMultiGetPastObjects(
		ctx context.Context,
		req iotaclient.TryMultiGetPastObjectsRequest,
	) ([]*iotajsonrpc.IotaPastObjectResponse, error)

	// Utility methods
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *iotago.Address,
		targetAmount uint64,
		gasAmount uint64,
	) (iotajsonrpc.Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		req *iotaclient.SignAndExecuteTransactionRequest,
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
	GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error)
	BatchGetObjectsOwnedByAddress(
		ctx context.Context,
		address *iotago.Address,
		options *iotajsonrpc.IotaObjectDataOptions,
		filterType string,
	) ([]iotajsonrpc.IotaObjectResponse, error)
	BatchGetFilteredObjectsOwnedByAddress(
		ctx context.Context,
		address *iotago.Address,
		options *iotajsonrpc.IotaObjectDataOptions,
		filter func(*iotajsonrpc.IotaObjectData) bool,
	) ([]iotajsonrpc.IotaObjectResponse, error)
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

var _ IotaClient = &iotaclient.Client{}
var _ IotaClient = &iotagraphql.GraphQLClient{}
