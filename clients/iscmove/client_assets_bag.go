package iscmove

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) AssetsBagNew(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*sui.ObjectRef, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "assets_bag",
				Function:      "new",
				TypeArguments: []sui.TypeTag{},
				Arguments:     []sui.Argument{},
			},
		},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}

	assetsBagRef, err := txnResponse.GetCreatedObjectInfo("assets_bag", "AssetsBag")
	if err != nil {
		return nil, fmt.Errorf("failed to create AssetsBag: %w", err)
	}

	return assetsBagRef, nil
}

func (c *Client) AssetsBagPlaceCoin(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	assetsBagRef *sui.ObjectRef,
	coin *sui.ObjectRef,
	coinType string,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TypeTag: %s: %w", coinType, err)
	}
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "assets_bag",
				Function:      "place_coin",
				TypeArguments: []sui.TypeTag{typeTag.Struct.TypeParams[0]},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: coin}),
				},
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

func (c *Client) AssetsDestroyEmpty(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	assetsBagRef *sui.ObjectRef,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "assets_bag",
				Function:      "destroy_empty",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
				},
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}

	return txnResponse, nil
}
