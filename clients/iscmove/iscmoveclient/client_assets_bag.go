package iscmoveclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func (c *Client) AssetsBagNew(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagNewAndTransfer(ptb, packageID, signer.Address())
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) AssetsBagPlaceCoin(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	assetsBagRef *iotago.ObjectRef,
	coin *iotago.ObjectRef,
	coinType iotajsonrpc.CoinType,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagPlaceCoin(
		ptb,
		packageID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coin}),
		string(coinType),
	)
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) AssetsBagPlaceCoinAmount(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	assetsBagRef *iotago.ObjectRef,
	coin *iotago.ObjectRef,
	coinType iotajsonrpc.CoinType,
	amount uint64,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagPlaceCoinWithAmount(ptb, packageID, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}), ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coin}), amount, string(coinType))
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) AssetsDestroyEmpty(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	assetsBagRef *iotago.ObjectRef,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsDestroyEmpty(ptb, packageID, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}))
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) AssetsBagTakeCoinBalanceMergeTo(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	assetsBagRef *iotago.ObjectRef,
	coinType iotajsonrpc.CoinType,
	amount uint64,
	mergeToCoin iotago.ObjectRef,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagTakeCoinBalanceMergeTo(
		ptb,
		packageID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
		amount,
		coinType,
	)
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) GetAssetsBagWithBalances(
	ctx context.Context,
	assetsBagID *iotago.ObjectID,
) (*iscmove.AssetsBagWithBalances, error) {
	fields, err := c.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{ParentObjectID: assetsBagID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in AssetsBag: %w", err)
	}

	bag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *assetsBagID,
			Size: uint64(len(fields.Data)),
		},
		Balances: make(iscmove.AssetsBagBalances),
	}
	for _, data := range fields.Data {
		resGetObject, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: &data.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowContent: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject for Balance: %w", err)
		}

		var moveBalance iotajsonrpc.MoveBalance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
		}

		cointype := iotajsonrpc.CoinType("0x" + data.Name.Value.(string))
		bag.Balances[cointype] = &iotajsonrpc.Balance{
			CoinType:     cointype,
			TotalBalance: iotajsonrpc.NewBigInt(moveBalance.Value.Uint64()),
		}
	}

	return &bag, nil
}
