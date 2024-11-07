package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// CreateAndSendRequest calls <packageID>::request::create_and_send_request() and transfers the created
// Request to the signer.
func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID iotago.PackageID,
	anchorAddress *iotago.ObjectID,
	assetsBagRef *iotago.ObjectRef,
	msg *iscmove.Message,
	allowance *iscmove.Assets,
	onchainGasBudget uint64,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)

	anchorRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: anchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBCreateAndSendRequest(
		ptb,
		packageID,
		*anchorRef.ObjectID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
		msg,
		allowance,
		onchainGasBudget,
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*iotago.ObjectRef{coinPage.Data[0].Ref()}
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
		signer,
		txnBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}

func (c *Client) CreateAndSendRequestWithAssets(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID iotago.PackageID,
	anchorAddress *iotago.ObjectID,
	assets *iscmove.Assets,
	msg *iscmove.Message,
	allowance *iscmove.Assets,
	onchainGasBudget uint64,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)

	anchorRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: anchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	getAllCoinsRes, err := c.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	placedCoins := []lo.Tuple2[*iotajsonrpc.Coin, uint64]{}
	// assume we can find it in the first page
	for cointype, bal := range assets.Coins {
		found := false
		for _, coin := range getAllCoinsRes.Data {
			assetsResource, err := iotago.NewResourceType(coin.CoinType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse resource type: %w", err)
			}
			getAllCoinsResource, err := iotago.NewResourceType(cointype)
			if err != nil {
				return nil, fmt.Errorf("failed to parse resource type: %w", err)
			}
			if assetsResource.String() == getAllCoinsResource.String() && coin.Balance.Uint64() > bal.Uint64() {
				placedCoins = append(placedCoins, lo.Tuple2[*iotajsonrpc.Coin, uint64]{A: coin, B: bal.Uint64()})
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot find coin for %s", cointype)
		}
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagNew(ptb, packageID, cryptolibSigner.Address())
	argAssetsBag := ptb.LastCommandResultArg()
	for _, tuple := range placedCoins {
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			packageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: tuple.A.Ref()}),
			tuple.B,
			tuple.A.CoinType,
		)
	}
	ptb = PTBCreateAndSendRequest(
		ptb,
		packageID,
		*anchorRef.ObjectID,
		argAssetsBag,
		msg,
		allowance,
		onchainGasBudget,
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		for _, coin := range coinPage.Data {
			if !pt.IsInInputObjects(coin.CoinObjectID) {
				gasPayments = []*iotago.ObjectRef{coin.Ref()}
				break
			}
		}
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
		signer,
		txnBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}

func (c *Client) GetRequestFromObjectID(
	ctx context.Context,
	reqID *iotago.ObjectID,
) (*iscmove.RefWithObject[iscmove.Request], error) {
	getObjectResponse, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: reqID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var req moveRequest
	err = iotaclient.UnmarshalBCS(getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	bals, err := c.GetAssetsBagWithBalances(context.Background(), &req.AssetsBag.Value.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AssetsBag of Request: %w", err)
	}
	req.AssetsBag.Value = bals
	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: getObjectResponse.Data.Ref(),
		Object:    req.ToRequest(),
	}, nil
}
