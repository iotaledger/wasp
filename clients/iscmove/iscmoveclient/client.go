package iscmoveclient

import (
	"context"
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/log"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*iotaclient.Client
	faucetURL string
}

func NewClient(client *iotaclient.Client, faucetURL string) *Client {
	return &Client{
		Client:    client,
		faucetURL: faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string, waitUntilEffectsVisible *iotaclient.WaitParams) *Client {
	return NewClient(
		iotaclient.NewHTTP(apiURL, waitUntilEffectsVisible),
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
	return NewClient(ws, faucetURL), nil
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
	var err error
	if len(gasPayments) == 0 {
		gasPayments, err = c.FindCoinsForGasPayment(
			ctx,
			signer.Address(),
			pt,
			gasPrice,
			gasBudget,
		)
		if err != nil {
			return nil, err
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
	var err error
	if len(gasPayments) == 0 {
		gasPayments, err = c.FindCoinsForGasPayment(
			ctx,
			signer.Address(),
			pt,
			gasPrice,
			gasBudget,
		)
		if err != nil {
			return nil, err
		}
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
