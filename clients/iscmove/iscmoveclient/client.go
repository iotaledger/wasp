package iscmoveclient

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	iotagraphql.IotaClient
	faucetURL string
}

func NewClient(iotaClient iotagraphql.IotaClient, faucetURL string) *Client {
	return &Client{
		IotaClient: iotaClient,
		faucetURL:  faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string, waitUntilEffectsVisible *iotagraphql.WaitParams) *Client {
	return NewClient(
		iotagraphql.NewGraphQLClientWithWaitParams(apiURL, waitUntilEffectsVisible),
		faucetURL,
	)
}

// NewWebsocketClient creates a new client. Note: websocket subscriptions are not
// currently supported, so this just creates an HTTP-based GraphQL client.
func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	waitUntilEffectsVisible *iotagraphql.WaitParams,
) (*Client, error) {
	_ = ctx
	return NewHTTPClient(wsURL, faucetURL, waitUntilEffectsVisible), nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	if c.faucetURL == "" {
		panic("missing faucetURL")
	}
	return iotagraphql.RequestFundsFromFaucet(ctx, address.AsIotaAddress(), c.faucetURL)
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
) (*iotagraphql.IotaTransactionBlockResponse, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasPrice, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		coins, err = iotagraphql.PickupCoinsWithFilter(
			coins,
			gasBudget,
			func(c *iotagraphql.Coin) bool { return !pt.IsInInputObjects(c.CoinObjectID) },
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
		&iotagraphql.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      signer,
			Options: &iotagraphql.IotaTransactionBlockResponseOptions{
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
) (*iotagraphql.DevInspectResults, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasPrice, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		coins, err = iotagraphql.PickupCoinsWithFilter(
			coins,
			gasBudget,
			func(c *iotagraphql.Coin) bool { return !pt.IsInInputObjects(c.CoinObjectID) },
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
		iotagraphql.DevInspectTransactionBlockRequest{
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

// WaitUntilStopped is a no-op placeholder. Websocket subscriptions are not currently supported.
func (c *Client) WaitUntilStopped() {}

// SubscribeEvent is not currently supported (websocket subscriptions not implemented).
func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *iotagraphql.IotaEventFilter,
	resultCh chan<- *iotagraphql.IotaEvent,
) error {
	return fmt.Errorf("SubscribeEvent is not supported: websocket subscriptions not implemented")
}

// SubscribeTransaction is not currently supported (websocket subscriptions not implemented).
func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *iotagraphql.TransactionFilter,
	resultCh chan<- *serialization.TagJson[iotagraphql.IotaTransactionBlockEffects],
) error {
	return fmt.Errorf("SubscribeTransaction is not supported: websocket subscriptions not implemented")
}

func (c *Client) GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error) {
	obj, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{ObjectID: &anchor, Options: &iotagraphql.IotaObjectDataOptions{
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
	txnBytes, err := c.Publish(ctx, iotagraphql.PublishRequest{
		Sender:          signer.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget * 10),
	})
	if err != nil {
		return iotago.PackageID{}, err
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		&iotagraphql.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotagraphql.IotaTransactionBlockResponseOptions{
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
