package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// CreateAndSendRequest calls <packageID>::request::create_and_send_request() and transfers the created
// Request to the signer.
func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	anchorAddress *sui.ObjectID,
	assetsBagRef *sui.ObjectRef,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	allowanceRef *sui.ObjectRef,
	onchainGasBudget uint64,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	anchorRes, err := c.GetObject(ctx, suiclient.GetObjectRequest{ObjectID: anchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb = PTBCreateAndSendRequest(
		ptb,
		packageID,
		*anchorRef.ObjectID,
		ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
		iscContractHname,
		iscFunctionHname,
		args,
		ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: allowanceRef}),
		onchainGasBudget,
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, suiclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*sui.ObjectRef{coinPage.Data[0].Ref()}
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	var txnBytes []byte
	if devMode {
		txnBytes, err = bcs.Marshal(tx.V1.Kind)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	} else {
		txnBytes, err = bcs.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}
