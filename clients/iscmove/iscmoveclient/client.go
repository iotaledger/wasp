package iscmoveclient

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/log"
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	client.IotaClient
	faucetURL string
	// FIXME this should be removed after we migrate to GqraphQL subscriptions
	wsClient *iotaclient.Client
}

func NewClient(iotaClient client.IotaClient, faucetURL string) *Client {
	return &Client{
		IotaClient: iotaClient,
		faucetURL:  faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string, waitUntilEffectsVisible *iotaclient.WaitParams) *Client {
	return NewClient(
		iotagraphql.NewGraphQLClientWithWaitParams(apiURL, waitUntilEffectsVisible),
		faucetURL,
	)
}

func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	waitUntilEffectsVisible *iotaclient.WaitParams,
	log log.Logger,
) (*Client, error) {
	ws, err := iotaclient.NewWebsocket(ctx, wsURL, waitUntilEffectsVisible, log)
	if err != nil {
		return nil, err
	}
	return &Client{
		IotaClient: ws,
		faucetURL:  faucetURL,
		wsClient:   ws,
	}, nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	if c.faucetURL == "" {
		panic("missing faucetURL")
	}
	return iotaclient.RequestFundsFromFaucet(ctx, address.AsIotaAddress(), c.faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestIotaSystemState(ctx)
	return err
}

func (c *Client) SignAndExecutePTB(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	pt iotago.ProgrammableTransaction,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasPrice, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		coins, err = iotajsonrpc.PickupCoinsWithFilter(
			coins,
			gasBudget,
			func(c *iotajsonrpc.Coin) bool { return !pt.IsInInputObjects(c.CoinObjectID) },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	if os.Getenv("DEBUG") != "" {
		pt.Print("-- SignAndExecutePTB -- ")
	}
	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:        true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}

func (c *Client) DevInspectPTB(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	pt iotago.ProgrammableTransaction,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.DevInspectResults, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasPrice, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		coins, err = iotajsonrpc.PickupCoinsWithFilter(
			coins,
			gasBudget,
			func(c *iotajsonrpc.Coin) bool { return !pt.IsInInputObjects(c.CoinObjectID) },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx.V1.Kind)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.DevInspectTransactionBlock(
		ctx,
		iotaclient.DevInspectTransactionBlockRequest{
			SenderAddress: signer.Address(),
			TxKindBytes:   txnBytes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if txnResponse.Error != "" {
		return nil, fmt.Errorf("execute error: %s", txnResponse.Error)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}

// WaitUntilStopped waits until the websocket client is stopped.
// This method is only available for websocket clients.
// FIXME this should be removed after we migrate to GqraphQL subscriptions
func (c *Client) WaitUntilStopped() {
	if c.wsClient == nil {
		panic("WaitUntilStopped is only available for websocket clients")
	}
	c.wsClient.WaitUntilStopped()
}

// SubscribeEvent subscribes to events matching the given filter.
// This method is only available for websocket clients.
// FIXME this should be removed after we migrate to GqraphQL subscriptions
func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *iotajsonrpc.EventFilter,
	resultCh chan<- *iotajsonrpc.IotaEvent,
) error {
	if c.wsClient == nil {
		return fmt.Errorf("SubscribeEvent is only available for websocket clients")
	}
	return c.wsClient.SubscribeEvent(ctx, filter, resultCh)
}

// SubscribeTransaction subscribes to transactions matching the given filter.
// This method is only available for websocket clients.
// FIXME this should be removed after we migrate to GqraphQL subscriptions
func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *iotajsonrpc.TransactionFilter,
	resultCh chan<- *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects],
) error {
	if c.wsClient == nil {
		return fmt.Errorf("SubscribeTransaction is only available for websocket clients")
	}
	return c.wsClient.SubscribeTransaction(ctx, filter, resultCh)
}

func (c *Client) GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error) {
	obj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: &anchor, Options: &iotajsonrpc.IotaObjectDataOptions{
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

	packageID := objectType.ResourceType().Address

	return *packageID, nil
}

func (c *Client) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
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
