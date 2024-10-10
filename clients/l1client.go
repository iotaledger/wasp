package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type L1Config struct {
	FaucetURL string
	APIURL    string
	GraphURL  string
}

type L1Client interface {
	GetDynamicFieldObject(
		ctx context.Context,
		req iotaclient.GetDynamicFieldObjectRequest,
	) (*iotajsonrpc.SuiObjectResponse, error)
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
	ResolveNameServiceAddress(ctx context.Context, suiName string) (*iotago.Address, error)
	ResolveNameServiceNames(
		ctx context.Context,
		req iotaclient.ResolveNameServiceNamesRequest,
	) (*iotajsonrpc.SuiNamePage, error)
	DevInspectTransactionBlock(
		ctx context.Context,
		req iotaclient.DevInspectTransactionBlockRequest,
	) (*iotajsonrpc.DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		txDataBytes iotago.Base64Data,
	) (*iotajsonrpc.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req iotaclient.ExecuteTransactionBlockRequest,
	) (*iotajsonrpc.SuiTransactionBlockResponse, error)
	GetCommitteeInfo(
		ctx context.Context,
		epoch *iotajsonrpc.BigInt, // optional
	) (*iotajsonrpc.CommitteeInfo, error)
	GetLatestSuiSystemState(ctx context.Context) (*iotajsonrpc.SuiSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error)
	GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error)
	GetStakesByIds(ctx context.Context, stakedSuiIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error)
	GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error)
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
	PayAllSui(
		ctx context.Context,
		req iotaclient.PayAllSuiRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	PaySui(
		ctx context.Context,
		req iotaclient.PaySuiRequest,
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
	TransferSui(
		ctx context.Context,
		req iotaclient.TransferSuiRequest,
	) (*iotajsonrpc.TransactionBytes, error)
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *iotago.Address,
		targetAmount uint64,
	) (iotajsonrpc.Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		signer iotasigner.Signer,
		txBytes iotago.Base64Data,
		options *iotajsonrpc.SuiTransactionBlockResponseOptions,
	) (*iotajsonrpc.SuiTransactionBlockResponse, error)
	PublishContract(
		ctx context.Context,
		signer iotasigner.Signer,
		modules []*iotago.Base64Data,
		dependencies []*iotago.Address,
		gasBudget uint64,
		options *iotajsonrpc.SuiTransactionBlockResponseOptions,
	) (*iotajsonrpc.SuiTransactionBlockResponse, *iotago.PackageID, error)
	MintToken(
		ctx context.Context,
		signer iotasigner.Signer,
		packageID *iotago.PackageID,
		tokenName string,
		treasuryCap *iotago.ObjectID,
		mintAmount uint64,
		options *iotajsonrpc.SuiTransactionBlockResponseOptions,
	) (*iotajsonrpc.SuiTransactionBlockResponse, error)
	GetSuiCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error)
	BatchGetObjectsOwnedByAddress(
		ctx context.Context,
		address *iotago.Address,
		options *iotajsonrpc.SuiObjectDataOptions,
		filterType string,
	) ([]iotajsonrpc.SuiObjectResponse, error)
	BatchGetFilteredObjectsOwnedByAddress(
		ctx context.Context,
		address *iotago.Address,
		options *iotajsonrpc.SuiObjectDataOptions,
		filter func(*iotajsonrpc.SuiObjectData) bool,
	) ([]iotajsonrpc.SuiObjectResponse, error)
	GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error)
	GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.SuiCoinMetadata, error)
	GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType iotago.ObjectType) (*iotajsonrpc.Supply, error)
	GetChainIdentifier(ctx context.Context) (string, error)
	GetCheckpoint(ctx context.Context, checkpointId *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error)
	GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error)
	GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.SuiEvent, error)
	GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error)
	GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.SuiObjectResponse, error)
	GetProtocolConfig(
		ctx context.Context,
		version *iotajsonrpc.BigInt, // optional
	) (*iotajsonrpc.ProtocolConfig, error)
	GetTotalTransactionBlocks(ctx context.Context) (string, error)
	GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.SuiTransactionBlockResponse, error)
	MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.SuiObjectResponse, error)
	MultiGetTransactionBlocks(
		ctx context.Context,
		req iotaclient.MultiGetTransactionBlocksRequest,
	) ([]*iotajsonrpc.SuiTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req iotaclient.TryGetPastObjectRequest,
	) (*iotajsonrpc.SuiPastObjectResponse, error)
	TryMultiGetPastObjects(
		ctx context.Context,
		req iotaclient.TryMultiGetPastObjectsRequest,
	) ([]*iotajsonrpc.SuiPastObjectResponse, error)
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error
	L2() L2Client
}

var _ L1Client = &l1Client{}

type l1Client struct {
	*iotaclient.Client

	Config L1Config
}

func (c *l1Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	faucetURL := c.Config.FaucetURL
	if faucetURL == "" {
		faucetURL = iotaconn.FaucetURL(c.Config.APIURL)
	}
	return iotaclient.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), faucetURL)
}

func (c *l1Client) Health(ctx context.Context) error {
	_, err := c.Client.GetLatestSuiSystemState(ctx)
	return err
}

func (c *l1Client) L2() L2Client {
	return iscmoveclient.NewClient(c.Client, c.Config.FaucetURL)
}

func NewL1Client(l1Config L1Config) L1Client {
	return &l1Client{
		iotaclient.NewHTTP(l1Config.APIURL),
		l1Config,
	}
}
