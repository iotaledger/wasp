package iscmoveclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) AllowanceNew(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb = PTBAllowanceNew(ptb, packageID, cryptolibSigner.Address())
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

func (c *Client) AllowanceWithCoinBalance(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	allowanceRef *sui.ObjectRef,
	coinBal uint64,
	coinType suijsonrpc.CoinType,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb = PTBAllowanceWithCoinBalance(ptb, packageID, ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: allowanceRef}), coinBal, string(coinType))
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

func (c *Client) AllowanceDestroy(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	allowanceRef *sui.ObjectRef,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb = PTBAllowanceDestroy(ptb, packageID, ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: allowanceRef}))
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

func (c *Client) GetAllowance(
	ctx context.Context,
	allowanceID *sui.ObjectID,
) (*iscmove.AllowanceWithVales, error) {
	fields, err := c.GetDynamicFields(ctx, suiclient.GetDynamicFieldsRequest{ParentObjectID: allowanceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in Allowance: %w", err)
	}

	allowance := iscmove.AllowanceWithVales{
		CoinAmounts: []uint64{},
		CoinTypes:   []suijsonrpc.CoinType{},
	}
	for _, data := range fields.Data {
		resGetObject, err := c.GetObject(ctx, suiclient.GetObjectRequest{
			ObjectID: &data.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowContent: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject for Allowance: %w", err)
		}

		var moveAllowance suijsonrpc.MoveAllowance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveAllowance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Allowance: %w", err)
		}

		cointype := suijsonrpc.CoinType("0x" + moveAllowance.Name.String())
		allowance.CoinAmounts = append(allowance.CoinAmounts, moveAllowance.Value.Uint64())
		allowance.CoinTypes = append(allowance.CoinTypes, cointype)
	}

	return &allowance, nil
}
