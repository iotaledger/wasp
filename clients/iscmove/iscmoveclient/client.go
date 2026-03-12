// Package iscmoveclient provides a client for interacting with ISC Move contracts.
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
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	iotagraphql.IotaClient
}

func NewClient(iotaClient iotagraphql.IotaClient) *Client {
	return &Client{
		IotaClient: iotaClient,
	}
}

// NewWebsocketClient creates a new client. Note: websocket subscriptions are not
// currently supported, so this just creates a GraphQL client.
func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	waitUntilEffectsVisible *iotagraphql.WaitParams,
) (*Client, error) {
	_ = ctx
	return NewClient(iotagraphql.NewGraphQLClientWithWaitParams(wsURL, faucetURL, waitUntilEffectsVisible)), nil
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
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	if len(gasPayments) > 0 {
		// Drop gas coins that already appear in PTB inputs or are duplicated.
		seen := map[iotago.ObjectID]struct{}{}
		filtered := make([]*iotago.ObjectRef, 0, len(gasPayments))
		for _, ref := range gasPayments {
			if ref == nil || ref.ObjectID == nil {
				continue
			}
			if pt.IsInInputObjects(ref.ObjectID) {
				continue
			}
			if _, ok := seen[*ref.ObjectID]; ok {
				continue
			}
			seen[*ref.ObjectID] = struct{}{}
			filtered = append(filtered, ref)
		}
		gasPayments = filtered
	}
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, *signer.Address(), gasPrice, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		coins, err = iotagraphql.PickupCoinsWithFilter(
			coins,
			gasBudget,
			func(c iotagraphql.Coin) bool { id := c.ObjectID(); return !pt.IsInInputObjects(&id) },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}
		gasPayments, err = coins.CoinRefs()
		if err != nil {
			return nil, fmt.Errorf("failed to get gas coin refs: %w", err)
		}
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
		txnBytes,
		signer,
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.ExecuteTransactionBlock.Effects.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.ExecuteTransactionBlock.Effects.GetErrors())
	}
	return txnResponse, nil
}

// WaitUntilStopped is a no-op placeholder. Websocket subscriptions are not currently supported.
func (c *Client) WaitUntilStopped() {}

func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *iotagraphql.IotaEventFilter,
	resultCh chan<- *iotagraphql.IotaEvent,
) error {
	return c.IotaClient.SubscribeEvent(ctx, filter, resultCh)
}

func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *iotagraphql.TransactionFilter,
	resultCh chan<- *iotagraphql.IotaTransactionBlockEffects,
) error {
	return c.IotaClient.SubscribeTransaction(ctx, filter, resultCh)
}

func (c *Client) GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error) {
	obj, err := c.GetObject(ctx, anchor)
	if err != nil {
		return iotago.PackageID{}, fmt.Errorf("retrieving anchor object: %w", err)
	}

	objectType, err := iotago.ObjectTypeFromString(obj.Object.TypeRepr())
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
		txnBytes.TxBytes,
		signer,
	)
	if err != nil {
		return iotago.PackageID{}, err
	}

	if !txnResponse.ExecuteTransactionBlock.Effects.IsSuccess() {
		return iotago.PackageID{}, errors.New("publish ISC contracts failed")
	}

	packageID := lo.Must(txnResponse.ExecuteTransactionBlock.Effects.GetPublishedPackageID())
	return *packageID, nil
}
