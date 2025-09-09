package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type L1Config struct {
	APIURL    string
	FaucetURL string
}

type L1Client interface {
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
		txDataBytes iotago.Base64Data,
	) (*iotajsonrpc.DryRunTransactionBlockResponse, error)
	ExecuteTransactionBlock(
		ctx context.Context,
		req iotaclient.ExecuteTransactionBlockRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetCommitteeInfo(
		ctx context.Context,
		epoch *iotajsonrpc.BigInt, // optional
	) (*iotajsonrpc.CommitteeInfo, error)
	GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error)
	GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error)
	GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error)
	GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error)
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
	PublishContract(
		ctx context.Context,
		signer iotasigner.Signer,
		modules []*iotago.Base64Data,
		dependencies []*iotago.Address,
		gasBudget uint64,
		options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error)
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
	GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error)
	GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error)
	GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error)
	GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error)
	GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error)
	GetChainIdentifier(ctx context.Context) (string, error)
	GetCheckpoint(ctx context.Context, checkpointID *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error)
	GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error)
	GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error)
	GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error)
	GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error)
	GetProtocolConfig(
		ctx context.Context,
		version *iotajsonrpc.BigInt, // optional
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
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error
	L2() L2Client
	IotaClient() *iotaclient.Client
	DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error)
	GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error)
	FindCoinsForGasPayment(
		ctx context.Context,
		owner *iotago.Address,
		pt iotago.ProgrammableTransaction,
		gasPrice uint64,
		gasBudget uint64,
	) ([]*iotago.ObjectRef, error)
	MergeCoinsAndExecute(
		ctx context.Context,
		owner iotasigner.Signer,
		destinationCoin *iotago.ObjectRef,
		sourceCoins []*iotago.ObjectRef,
		gasBudget uint64,
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

	WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error)
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
	return iotaclient.RequestFundsFromFaucet(ctx, address.AsIotaAddress(), faucetURL)
}

func (c *l1Client) Health(ctx context.Context) error {
	_, err := c.GetLatestIotaSystemState(ctx)
	return err
}

func (c *l1Client) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	iscBytecode := contracts.ISC()
	txnBytes, err := c.Publish(ctx, iotaclient.PublishRequest{
		Sender:          signer.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	if err != nil {
		return iotago.PackageID{}, err
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	if err != nil {
		return iotago.PackageID{}, err
	}

	if !txnResponse.Effects.Data.IsSuccess() {
		return iotago.PackageID{}, errors.New("publish ISC contracts failed")
	}
	packageID := lo.Must(txnResponse.GetPublishedPackageID())
	return *packageID, nil
}

func (c *l1Client) GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error) {
	obj, err := c.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: &anchor, Options: &iotajsonrpc.IotaObjectDataOptions{
		ShowDisplay: true,
		ShowType:    true,
	}})
	if err != nil {
		return iotago.PackageID{}, fmt.Errorf("retrieving anchor object: %w", err)
	}

	objectType, err := iotago.ObjectTypeFromString(*obj.Data.Type)
	if err != nil {
		return iotago.PackageID{}, fmt.Errorf("parsing anchor object type: %w", err)
	}

	addr := objectType.ResourceType().Address
	packageID, err := iotago.PackageIDFromHex(addr.String())
	if err != nil {
		return iotago.PackageID{}, fmt.Errorf("parsing package ID: %w", err)
	}

	return *packageID, nil
}

func (c *l1Client) L2() L2Client {
	return iscmoveclient.NewClient(c.Client, c.Config.FaucetURL)
}

func (c *l1Client) IotaClient() *iotaclient.Client {
	return c.Client
}

// WaitForNextVersionForTesting waits for an object to change its version.
// This tries to make sure that an object meant to be used multiple times, does not get referenced twice with the same ref.
// Handle with care. Only use it on objects that are expected to be used again, like a GasCoin/Generic coin/Requests
func (c *l1Client) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	// Some 'sugar' to make dynamic refs handling easier (where refs can be nil or set depending on state)
	if currentRef == nil {
		cb()
		return currentRef, nil
	}

	cb()

	// Create a ticker for polling
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	// Add timeout to context if not already set
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("WaitForNextVersionForTesting: context deadline exceeded while waiting for object version change: %v", currentRef)
		case <-ticker.C:
			// Poll for object update
			newRef, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: currentRef.ObjectID})
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
				}
				continue
			}

			if newRef.Error != nil {
				// The provided object got consumed and is gone. We can return.
				if newRef.Error.Data.Deleted != nil || newRef.Error.Data.NotExists != nil {
					return currentRef, nil
				}

				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: object error: %v, retrying...", newRef.Error)
				}
				continue
			}

			if newRef.Data.Ref().Version > currentRef.Version {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, newRef.Data.Ref())
				}

				ref := newRef.Data.Ref()
				return &ref, nil
			}

			if logger != nil {
				logger.LogInfof("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			}
		}
	}
}

func NewL1Client(l1Config L1Config, waitUntilEffectsVisible *iotaclient.WaitParams) L1Client {
	return &l1Client{
		iotaclient.NewHTTP(l1Config.APIURL, waitUntilEffectsVisible),
		l1Config,
	}
}

func NewLocalnetClient(waitUntilEffectsVisible *iotaclient.WaitParams) L1Client {
	return NewL1Client(L1Config{
		APIURL:    iotaconn.LocalnetEndpointURL,
		FaucetURL: iotaconn.LocalnetFaucetURL,
	}, waitUntilEffectsVisible)
}
