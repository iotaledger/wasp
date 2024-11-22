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
				ShowEffects:       true,
				ShowObjectChanges: true,
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

// CreateAndSendRequest calls <packageID>::request::create_and_send_request() and transfers the created
// Request to the signer.
func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	signer cryptolib.Signer,
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
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) CreateAndSendRequestWithAssets(
	ctx context.Context,
	signer cryptolib.Signer,
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
	anchorRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: anchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	allCoins, err := c.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	placedCoins := []lo.Tuple2[*iotajsonrpc.Coin, uint64]{}
	// assume we can find it in the first page
	for cointype, bal := range assets.Coins {
		coin, ok := lo.Find(allCoins.Data, func(coin *iotajsonrpc.Coin) bool {
			if !lo.Must(iotago.IsSameResource(cointype, coin.CoinType)) {
				return false
			}
			if lo.ContainsBy(gasPayments, func(ref *iotago.ObjectRef) bool {
				return ref.ObjectID.Equals(*coin.CoinObjectID)
			}) {
				return false
			}
			return coin.Balance.Uint64() >= bal.Uint64()
		})
		if !ok {
			return nil, fmt.Errorf("cannot find coin for type %s", cointype)
		}
		placedCoins = append(placedCoins, lo.Tuple2[*iotajsonrpc.Coin, uint64]{A: coin, B: bal.Uint64()})
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagNew(ptb, packageID, signer.Address())
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
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
}

func (c *Client) GetRequestFromObjectID(
	ctx context.Context,
	reqID *iotago.ObjectID,
) (*iscmove.RefWithObject[iscmove.Request], error) {
	getObjectResponse, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: reqID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get request content: %w", err)
	}
	if getObjectResponse.Data == nil {
		return nil, fmt.Errorf("request %s not found", *reqID)
	}
	return c.parseRequestAndFetchAssetsBag(getObjectResponse.Data)
}

func (c *Client) parseRequestAndFetchAssetsBag(obj *iotajsonrpc.IotaObjectData) (*iscmove.RefWithObject[iscmove.Request], error) {
	var req moveRequest
	err := iotaclient.UnmarshalBCS(obj.Bcs.Data.MoveObject.BcsBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	bals, err := c.GetAssetsBagWithBalances(context.Background(), &req.AssetsBag.Value.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AssetsBag of Request: %w", err)
	}
	req.AssetsBag.Value = bals
	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: obj.Ref(),
		Object:    req.ToRequest(),
		Owner:     obj.Owner.AddressOwner,
	}, nil
}

func (c *Client) GetRequests(
	ctx context.Context,
	packageID iotago.PackageID,
	anchorAddress *iotago.ObjectID,
) (
	[]*iscmove.RefWithObject[iscmove.Request],
	error,
) {
	reqs := make([]*iscmove.RefWithObject[iscmove.Request], 0)
	var lastSeen *iotago.ObjectID
	for {
		res, err := c.GetOwnedObjects(ctx, iotaclient.GetOwnedObjectsRequest{
			Address: anchorAddress,
			Query: &iotajsonrpc.IotaObjectResponseQuery{
				Filter: &iotajsonrpc.IotaObjectDataFilter{
					StructType: &iotago.StructTag{
						Address: &packageID,
						Module:  iscmove.RequestModuleName,
						Name:    iscmove.RequestObjectName,
					},
				},
				Options: &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
			},
			Cursor: lastSeen,
		})
		if ctx.Err() != nil {
			return nil, fmt.Errorf("failed to fetch requests: %w", err)
		}
		if len(res.Data) == 0 {
			break
		}
		lastSeen = res.NextCursor
		for _, reqData := range res.Data {
			req, err := c.parseRequestAndFetchAssetsBag(reqData.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode request: %w", err)
			}
			reqs = append(reqs, req)
		}
	}
	return reqs, nil
}
