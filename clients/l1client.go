package clients

import (
	"context"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

type L1Config struct {
	FaucetURL string
	APIURL    string
}

type L1Client interface {
	GetDynamicFieldObject(
		ctx context.Context,
		req suiclient.GetDynamicFieldObjectRequest,
	) (*suijsonrpc.SuiObjectResponse, error)
	GetDynamicFields(
		ctx context.Context,
		req suiclient.GetDynamicFieldsRequest,
	) (*suijsonrpc.DynamicFieldPage, error)
	GetOwnedObjects(
		ctx context.Context,
		req suiclient.GetOwnedObjectsRequest,
	) (*suijsonrpc.ObjectsPage, error)
	QueryEvents(
		ctx context.Context,
		req suiclient.QueryEventsRequest,
	) (*suijsonrpc.EventPage, error)
	QueryTransactionBlocks(
		ctx context.Context,
		req suiclient.QueryTransactionBlocksRequest,
	) (*suijsonrpc.TransactionBlocksPage, error)
	ResolveNameServiceAddress(ctx context.Context, suiName string) (*sui.Address, error)
	ResolveNameServiceNames(
		ctx context.Context,
		req suiclient.ResolveNameServiceNamesRequest,
	) (*suijsonrpc.SuiNamePage, error)
	SubscribeEvent(
		ctx context.Context,
		filter *suijsonrpc.EventFilter,
		resultCh chan suijsonrpc.SuiEvent,
	) error
	DevInspectTransactionBlock(
		ctx context.Context,
		req suiclient.DevInspectTransactionBlockRequest,
	) (*suijsonrpc.DevInspectResults, error)
	DryRunTransaction(
		ctx context.Context,
		txDataBytes sui.Base64Data,
	) (*suijsonrpc.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req suiclient.ExecuteTransactionBlockRequest,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	GetCommitteeInfo(
		ctx context.Context,
		epoch *suijsonrpc.BigInt, // optional

	) (*suijsonrpc.CommitteeInfo, error)
	GetLatestSuiSystemState(ctx context.Context) (*suijsonrpc.SuiSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*suijsonrpc.BigInt, error)
	GetStakes(ctx context.Context, owner *sui.Address) ([]*suijsonrpc.DelegatedStake, error)
	GetStakesByIds(ctx context.Context, stakedSuiIds []sui.ObjectID) ([]*suijsonrpc.DelegatedStake, error)
	GetValidatorsApy(ctx context.Context) (*suijsonrpc.ValidatorsApy, error)
	BatchTransaction(
		ctx context.Context,
		req suiclient.BatchTransactionRequest,
	) (*suijsonrpc.TransactionBytes, error)
	MergeCoins(
		ctx context.Context,
		req suiclient.MergeCoinsRequest,
	) (*suijsonrpc.TransactionBytes, error)
	MoveCall(
		ctx context.Context,
		req suiclient.MoveCallRequest,
	) (*suijsonrpc.TransactionBytes, error)
	Pay(
		ctx context.Context,
		req suiclient.PayRequest,
	) (*suijsonrpc.TransactionBytes, error)
	PayAllSui(
		ctx context.Context,
		req suiclient.PayAllSuiRequest,
	) (*suijsonrpc.TransactionBytes, error)
	PaySui(
		ctx context.Context,
		req suiclient.PaySuiRequest,
	) (*suijsonrpc.TransactionBytes, error)
	Publish(
		ctx context.Context,
		req suiclient.PublishRequest,
	) (*suijsonrpc.TransactionBytes, error)
	RequestAddStake(
		ctx context.Context,
		req suiclient.RequestAddStakeRequest,
	) (*suijsonrpc.TransactionBytes, error)
	RequestWithdrawStake(
		ctx context.Context,
		req suiclient.RequestWithdrawStakeRequest,
	) (*suijsonrpc.TransactionBytes, error)
	SplitCoin(
		ctx context.Context,
		req suiclient.SplitCoinRequest,
	) (*suijsonrpc.TransactionBytes, error)
	SplitCoinEqual(
		ctx context.Context,
		req suiclient.SplitCoinEqualRequest,
	) (*suijsonrpc.TransactionBytes, error)
	TransferObject(
		ctx context.Context,
		req suiclient.TransferObjectRequest,
	) (*suijsonrpc.TransactionBytes, error)
	TransferSui(
		ctx context.Context,
		req suiclient.TransferSuiRequest,
	) (*suijsonrpc.TransactionBytes, error)
	GetCoinObjsForTargetAmount(
		ctx context.Context,
		address *sui.Address,
		targetAmount uint64,
	) (suijsonrpc.Coins, error)
	SignAndExecuteTransaction(
		ctx context.Context,
		signer suisigner.Signer,
		txBytes sui.Base64Data,
		options *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	PublishContract(
		ctx context.Context,
		signer suisigner.Signer,
		modules []*sui.Base64Data,
		dependencies []*sui.Address,
		gasBudget uint64,
		options *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, *sui.PackageID, error)
	MintToken(
		ctx context.Context,
		signer suisigner.Signer,
		packageID *sui.PackageID,
		tokenName string,
		treasuryCap *sui.ObjectID,
		mintAmount uint64,
		options *suijsonrpc.SuiTransactionBlockResponseOptions,
	) (*suijsonrpc.SuiTransactionBlockResponse, error)
	GetSuiCoinsOwnedByAddress(ctx context.Context, address *sui.Address) (suijsonrpc.Coins, error)
	BatchGetObjectsOwnedByAddress(
		ctx context.Context,
		address *sui.Address,
		options *suijsonrpc.SuiObjectDataOptions,
		filterType string,
	) ([]suijsonrpc.SuiObjectResponse, error)
	BatchGetFilteredObjectsOwnedByAddress(
		ctx context.Context,
		address *sui.Address,
		options *suijsonrpc.SuiObjectDataOptions,
		filter func(*suijsonrpc.SuiObjectData) bool,
	) ([]suijsonrpc.SuiObjectResponse, error)
	GetAllBalances(ctx context.Context, owner *sui.Address) ([]*suijsonrpc.Balance, error)
	GetAllCoins(ctx context.Context, req suiclient.GetAllCoinsRequest) (*suijsonrpc.CoinPage, error)
	GetBalance(ctx context.Context, req suiclient.GetBalanceRequest) (*suijsonrpc.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*suijsonrpc.SuiCoinMetadata, error)
	GetCoins(ctx context.Context, req suiclient.GetCoinsRequest) (*suijsonrpc.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType sui.ObjectType) (*suijsonrpc.Supply, error)
	GetChainIdentifier(ctx context.Context) (string, error)
	GetCheckpoint(ctx context.Context, checkpointId *suijsonrpc.BigInt) (*suijsonrpc.Checkpoint, error)
	GetCheckpoints(ctx context.Context, req suiclient.GetCheckpointsRequest) (*suijsonrpc.CheckpointPage, error)
	GetEvents(ctx context.Context, digest *sui.TransactionDigest) ([]*suijsonrpc.SuiEvent, error)
	GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error)
	GetObject(ctx context.Context, req suiclient.GetObjectRequest) (*suijsonrpc.SuiObjectResponse, error)
	GetProtocolConfig(
		ctx context.Context,
		version *suijsonrpc.BigInt, // optional
	) (*suijsonrpc.ProtocolConfig, error)
	GetTotalTransactionBlocks(ctx context.Context) (string, error)
	GetTransactionBlock(ctx context.Context, req suiclient.GetTransactionBlockRequest) (*suijsonrpc.SuiTransactionBlockResponse, error)
	MultiGetObjects(ctx context.Context, req suiclient.MultiGetObjectsRequest) ([]suijsonrpc.SuiObjectResponse, error)
	MultiGetTransactionBlocks(
		ctx context.Context,
		req suiclient.MultiGetTransactionBlocksRequest,
	) ([]*suijsonrpc.SuiTransactionBlockResponse, error)
	TryGetPastObject(
		ctx context.Context,
		req suiclient.TryGetPastObjectRequest,
	) (*suijsonrpc.SuiPastObjectResponse, error)
	TryMultiGetPastObjects(
		ctx context.Context,
		req suiclient.TryMultiGetPastObjectsRequest,
	) ([]*suijsonrpc.SuiPastObjectResponse, error)
	WithSignerAndFund(seed []byte, index int) (*suiclient.Client, suisigner.Signer)
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error
	WithWebsocket(url string)
}

var _ L1Client = &L1ClientExt{}

type L1ClientExt struct {
	*suiclient.Client

	Config L1Config
}

func (c *L1ClientExt) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	faucetURL := c.Config.FaucetURL

	if faucetURL == "" {
		switch c.Config.APIURL {
		case suiconn.TestnetEndpointURL:
			faucetURL = suiconn.TestnetFaucetURL
		case suiconn.DevnetEndpointURL:
			faucetURL = suiconn.DevnetFaucetURL
		case suiconn.LocalnetEndpointURL:
			faucetURL = suiconn.LocalnetFaucetURL
		default:
			panic("unspecified FaucetURL")
		}
	}

	return suiclient.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), faucetURL)
}

func (c *L1ClientExt) Health(ctx context.Context) error {
	_, err := c.Client.GetLatestSuiSystemState(ctx)
	return err
}

func NewL1Client(l1Config L1Config) L1Client {
	return &L1ClientExt{
		suiclient.New(l1Config.APIURL),
		l1Config,
	}
}
