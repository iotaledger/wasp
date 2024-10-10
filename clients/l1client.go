package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
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
		req suiclient2.GetDynamicFieldObjectRequest,
	) (*suijsonrpc2.SuiObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req suiclient2.GetDynamicFieldsRequest,
	) (*suijsonrpc2.DynamicFieldPage, error)
	GetOwnedObjects(
		ctx context.Context,
		req suiclient2.GetOwnedObjectsRequest,
	) (*suijsonrpc2.ObjectsPage, error)
	QueryEvents(
		ctx context.Context,
		req suiclient2.QueryEventsRequest,
	) (*suijsonrpc2.EventPage, error)
	QueryTransactionBlocks(
		ctx context.Context,
		req suiclient2.QueryTransactionBlocksRequest,
	) (*suijsonrpc2.TransactionBlocksPage, error)
	ResolveNameServiceAddress(ctx context.Context, suiName string) (*sui2.Address, error)
	ResolveNameServiceNames(
		ctx context.Context,
		req suiclient2.ResolveNameServiceNamesRequest,
	) (*suijsonrpc2.SuiNamePage, error)
	DevInspectTransactionBlock(
		ctx context.Context,
		req suiclient2.DevInspectTransactionBlockRequest,
	) (*suijsonrpc2.DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		txDataBytes sui2.Base64Data,
	) (*suijsonrpc2.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req suiclient2.ExecuteTransactionBlockRequest,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	GetCommitteeInfo(
		ctx context.Context,
		epoch *suijsonrpc2.BigInt, // optional
	) (*suijsonrpc2.CommitteeInfo, error)
	GetLatestSuiSystemState(ctx context.Context) (*suijsonrpc2.SuiSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*suijsonrpc2.BigInt, error)
	GetStakes(ctx context.Context, owner *sui2.Address) ([]*suijsonrpc2.DelegatedStake, error)
	GetStakesByIds(ctx context.Context, stakedSuiIds []sui2.ObjectID) ([]*suijsonrpc2.DelegatedStake, error)
	GetValidatorsApy(ctx context.Context) (*suijsonrpc2.ValidatorsApy, error)
	BatchTransaction(
		ctx context.Context,
		req suiclient2.BatchTransactionRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	MergeCoins(
		ctx context.Context,
		req suiclient2.MergeCoinsRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	MoveCall(
		ctx context.Context,
		req suiclient2.MoveCallRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	Pay(
		ctx context.Context,
		req suiclient2.PayRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	PayAllSui(
		ctx context.Context,
		req suiclient2.PayAllSuiRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	PaySui(
		ctx context.Context,
		req suiclient2.PaySuiRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	Publish(
		ctx context.Context,
		req suiclient2.PublishRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	RequestAddStake(
		ctx context.Context,
		req suiclient2.RequestAddStakeRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	RequestWithdrawStake(
		ctx context.Context,
		req suiclient2.RequestWithdrawStakeRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	SplitCoin(
		ctx context.Context,
		req suiclient2.SplitCoinRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	SplitCoinEqual(
		ctx context.Context,
		req suiclient2.SplitCoinEqualRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req suiclient2.TransferObjectRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	TransferSui(
		ctx context.Context,
		req suiclient2.TransferSuiRequest,
	) (*suijsonrpc2.TransactionBytes, error)
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *sui2.Address,
		targetAmount uint64,
	) (suijsonrpc2.Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		signer suisigner.Signer,
		txBytes sui2.Base64Data,
		options *suijsonrpc2.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	PublishContract(
		ctx context.Context,
		signer suisigner.Signer,
		modules []*sui2.Base64Data,
		dependencies []*sui2.Address,
		gasBudget uint64,
		options *suijsonrpc2.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc2.SuiTransactionBlockResponse, *sui2.PackageID, error)
	MintToken(
		ctx context.Context,
		signer suisigner.Signer,
		packageID *sui2.PackageID,
		tokenName string,
		treasuryCap *sui2.ObjectID,
		mintAmount uint64,
		options *suijsonrpc2.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	GetSuiCoinsOwnedByAddress(ctx context.Context, address *sui2.Address) (suijsonrpc2.Coins, error)
	BatchGetObjectsOwnedByAddress(
		ctx context.Context,
		address *sui2.Address,
		options *suijsonrpc2.SuiObjectDataOptions,
		filterType string,
	) ([]suijsonrpc2.SuiObjectResponse, error)
	BatchGetFilteredObjectsOwnedByAddress(
		ctx context.Context,
		address *sui2.Address,
		options *suijsonrpc2.SuiObjectDataOptions,
		filter func(*suijsonrpc2.SuiObjectData) bool,
	) ([]suijsonrpc2.SuiObjectResponse, error)
	GetAllBalances(ctx context.Context, owner *sui2.Address) ([]*suijsonrpc2.Balance, error)
	GetAllCoins(ctx context.Context, req suiclient2.GetAllCoinsRequest) (*suijsonrpc2.CoinPage, error)
	GetBalance(ctx context.Context, req suiclient2.GetBalanceRequest) (*suijsonrpc2.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*suijsonrpc2.SuiCoinMetadata, error)
	GetCoins(ctx context.Context, req suiclient2.GetCoinsRequest) (*suijsonrpc2.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType sui2.ObjectType) (*suijsonrpc2.Supply, error)
	GetChainIdentifier(ctx context.Context) (string, error)
	GetCheckpoint(ctx context.Context, checkpointId *suijsonrpc2.BigInt) (*suijsonrpc2.Checkpoint, error)
	GetCheckpoints(ctx context.Context, req suiclient2.GetCheckpointsRequest) (*suijsonrpc2.CheckpointPage, error)
	GetEvents(ctx context.Context, digest *sui2.TransactionDigest) ([]*suijsonrpc2.SuiEvent, error)
	GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error)
	GetObject(ctx context.Context, req suiclient2.GetObjectRequest) (*suijsonrpc2.SuiObjectResponse, error)
	GetProtocolConfig(
		ctx context.Context,
		version *suijsonrpc2.BigInt, // optional
	) (*suijsonrpc2.ProtocolConfig, error)
	GetTotalTransactionBlocks(ctx context.Context) (string, error)
	GetTransactionBlock(ctx context.Context, req suiclient2.GetTransactionBlockRequest) (*suijsonrpc2.SuiTransactionBlockResponse, error)
	MultiGetObjects(ctx context.Context, req suiclient2.MultiGetObjectsRequest) ([]suijsonrpc2.SuiObjectResponse, error)
	MultiGetTransactionBlocks(
		ctx context.Context,
		req suiclient2.MultiGetTransactionBlocksRequest,
	) ([]*suijsonrpc2.SuiTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req suiclient2.TryGetPastObjectRequest,
	) (*suijsonrpc2.SuiPastObjectResponse, error)
	TryMultiGetPastObjects(
		ctx context.Context,
		req suiclient2.TryMultiGetPastObjectsRequest,
	) ([]*suijsonrpc2.SuiPastObjectResponse, error)
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error
	L2() L2Client
}

var _ L1Client = &l1Client{}

type l1Client struct {
	*suiclient2.Client

	Config L1Config
}

func (c *l1Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	faucetURL := c.Config.FaucetURL
	if faucetURL == "" {
		faucetURL = suiconn.FaucetURL(c.Config.APIURL)
	}
	return suiclient2.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), faucetURL)
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
		suiclient2.NewHTTP(l1Config.APIURL),
		l1Config,
	}
}
