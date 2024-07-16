package iscmove

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) AssetsBagNew(
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

	ptb := NewAssetsBagNewPTB(packageID, cryptolibSigner.Address())

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		ptb,
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

func (c *Client) AssetsBagPlaceCoin(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	assetsBagRef *sui.ObjectRef,
	coin *sui.ObjectRef,
	coinType suijsonrpc.CoinType,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb, err := NewAssetsBagPlaceCoinPTB(packageID, assetsBagRef, coin, string(coinType))
	if err != nil {
		return nil, fmt.Errorf("can't create PTB: %w", err)
	}

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		ptb,
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

func (c *Client) AssetsDestroyEmpty(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	assetsBagRef *sui.ObjectRef,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := NewAssetsDestroyEmptyPTB(packageID, assetsBagRef)

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		ptb,
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

func (c *Client) GetAssetsBagFromAssetsBagID(ctx context.Context, assetsBagObjectID *sui.ObjectID) (*AssetsBag, error) {
	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: assetsBagObjectID,
		Options: &suijsonrpc.SuiObjectDataOptions{
			ShowContent: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject for AssetsBag: %w", err)
	}

	var tmpMoveAssetsBag MoveAssetsBag
	err = json.Unmarshal(getObjectResponse.Data.Content.Data.MoveObject.Fields, &tmpMoveAssetsBag)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields in AssetsBag: %w", err)
	}

	assetBag := NewAssetsBag()
	assetBag.ID = *tmpMoveAssetsBag.ID.ID
	assetBag.Size = tmpMoveAssetsBag.Size.Uint64()

	fields, err := c.GetDynamicFields(ctx, suiclient.GetDynamicFieldsRequest{ParentObjectID: assetsBagObjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in AssetsBag: %w", err)
	}

	for _, data := range fields.Data {
		resGetObject, err := c.GetObject(ctx, suiclient.GetObjectRequest{
			ObjectID: &data.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowContent: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject for Balance: %w", err)
		}

		var moveBalance suijsonrpc.MoveBalance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
		}

		cointype := suijsonrpc.CoinType("0x" + data.Name.Value.(string))
		assetBag.Balances[cointype] = &suijsonrpc.Balance{
			CoinType:     cointype,
			TotalBalance: moveBalance.Value,
		}
	}

	return assetBag, nil
}

func (c *Client) GetAssetsBagFromAnchor(ctx context.Context, anchorObjID *sui.ObjectID) (*AssetsBag, error) {
	getObjectResponseAnchor, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjID,
		Options: &suijsonrpc.SuiObjectDataOptions{
			ShowContent: true,
			ShowBcs:     true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject for Anchor: %w", err)
	}

	var tmpAnchorJsonObject anchorJsonObject
	err = json.Unmarshal(getObjectResponseAnchor.Data.Content.Data.MoveObject.Fields, &tmpAnchorJsonObject)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields in Anchor: %w", err)
	}

	assetBag := NewAssetsBag()
	assetBag.ID = *tmpAnchorJsonObject.Assets.Fields.Value.Fields.ID.ID
	assetBag.Size = tmpAnchorJsonObject.Assets.Fields.Value.Fields.Size.Uint64()

	fields, err := c.GetDynamicFields(ctx, suiclient.GetDynamicFieldsRequest{ParentObjectID: tmpAnchorJsonObject.Assets.Fields.Value.Fields.ID.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in Anchor: %w", err)
	}

	for _, data := range fields.Data {
		resGetObject, err := c.GetObject(ctx, suiclient.GetObjectRequest{
			ObjectID: &data.ObjectID,
			Options: &suijsonrpc.SuiObjectDataOptions{
				ShowContent: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject for Balance: %w", err)
		}

		var moveBalance suijsonrpc.MoveBalance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
		}
		cointype := suijsonrpc.CoinType("0x" + data.Name.Value.(string))
		assetBag.Balances[cointype] = &suijsonrpc.Balance{
			CoinType:     cointype,
			TotalBalance: moveBalance.Value,
		}
	}

	return assetBag, nil
}

func (c *Client) GetAssetsBagFromRequestID(ctx context.Context, requestObjID *sui.ObjectID) (*AssetsBag, error) {
	getObjectResponseRequest, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: requestObjID,
		Options: &suijsonrpc.SuiObjectDataOptions{
			ShowContent: true,
			ShowBcs:     true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject for Request: %w", err)
	}

	var tmpRequestJsonObject requestJsonObject
	err = json.Unmarshal(getObjectResponseRequest.Data.Content.Data.MoveObject.Fields, &tmpRequestJsonObject)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields in Request: %w", err)
	}

	assetBag := NewAssetsBag()
	assetBag.ID = *tmpRequestJsonObject.AssetsBag.Fields.Value.Fields.ID.ID
	assetBag.Size = tmpRequestJsonObject.AssetsBag.Fields.Value.Fields.Size.Uint64()

	fields, err := c.GetDynamicFields(ctx, suiclient.GetDynamicFieldsRequest{ParentObjectID: tmpRequestJsonObject.AssetsBag.Fields.Value.Fields.ID.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in Request: %w", err)
	}

	for _, data := range fields.Data {
		resGetObject, err := c.GetObject(ctx, suiclient.GetObjectRequest{
			ObjectID: &data.ObjectID,
			Options: &suijsonrpc.SuiObjectDataOptions{
				ShowContent: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject for Balance: %w", err)
		}

		var moveBalance suijsonrpc.MoveBalance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
		}
		cointype := suijsonrpc.CoinType("0x" + data.Name.Value.(string))
		assetBag.Balances[cointype] = &suijsonrpc.Balance{
			CoinType:     cointype,
			TotalBalance: moveBalance.Value,
		}
	}

	return assetBag, nil
}
