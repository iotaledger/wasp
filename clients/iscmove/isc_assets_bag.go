package iscmove

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// TODO add coin/balance/assets as parameters.
func (c *Client) AssetsBagNew(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	arg1 := ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "assets_bag",
				Function:      "new",
				TypeArguments: []sui_types.TypeTag{},
				Arguments:     []sui_types.Argument{},
			},
		},
	)
	ptb.Command(
		sui_types.Command{
			TransferObjects: &sui_types.ProgrammableTransferObjects{
				Objects: []sui_types.Argument{arg1},
				Address: ptb.MustPure(signer.Address),
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address,
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

// A single function covers `place_coin()`, `place_coin_balance()`, `place_asset()`
func (c *Client) AssetsBagAddItems(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	assetsBagRef *sui_types.ObjectRef,
	coins []*sui_types.ObjectRef,
	coinBalances []*sui_types.ObjectRef,
	assets []*sui_types.ObjectRef,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()
	argAssetsBag := ptb.MustObj(sui_types.ObjectArg{ImmOrOwnedObject: assetsBagRef})
	if len(coins) != 0 {
		for _, obj := range coins {
			ptb.Command(
				sui_types.Command{
					MoveCall: &sui_types.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "assets_bag",
						Function:      "place_coin",
						TypeArguments: []sui_types.TypeTag{},
						Arguments:     []sui_types.Argument{argAssetsBag, ptb.MustObj(sui_types.ObjectArg{ImmOrOwnedObject: obj})},
					},
				},
			)
		}
	}
	if len(coinBalances) != 0 {
		for _, obj := range coinBalances {
			ptb.Command(
				sui_types.Command{
					MoveCall: &sui_types.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "assets_bag",
						Function:      "place_coin_balance",
						TypeArguments: []sui_types.TypeTag{},
						Arguments:     []sui_types.Argument{argAssetsBag, ptb.MustObj(sui_types.ObjectArg{ImmOrOwnedObject: obj})},
					},
				},
			)
		}
	}
	if len(coinBalances) != 0 {
		for _, obj := range assets {
			ptb.Command(
				sui_types.Command{
					MoveCall: &sui_types.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "assets_bag",
						Function:      "place_asset",
						TypeArguments: []sui_types.TypeTag{},
						Arguments:     []sui_types.Argument{argAssetsBag, ptb.MustObj(sui_types.ObjectArg{ImmOrOwnedObject: obj})},
					},
				},
			)
		}
	}
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address,
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	// txnBytes, err := bcs.Marshal(tx)
	// if err != nil {
	// 	return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	// }

	// txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	// if err != nil {
	// 	return nil, fmt.Errorf("can't execute the transaction: %w", err)
	// }

	txnBytes, err := bcs.Marshal(tx.V1.Kind)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.DevInspectTransactionBlock(ctx, signer.Address, txnBytes, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	fmt.Println("err: ", err)
	fmt.Println("txnResponse: ", txnResponse.Effects.Data.V1.Status.Error)

	return nil, nil
}
