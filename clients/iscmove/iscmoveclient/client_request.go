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
)

type CreateAndSendRequestRequest struct {
	Signer           cryptolib.Signer
	PackageID        iotago.PackageID
	AnchorAddress    *iotago.ObjectID
	AssetsBagRef     *iotago.ObjectRef
	Message          *iscmove.Message
	Allowance        *iscmove.Assets
	OnchainGasBudget uint64
	GasPayments      []*iotago.ObjectRef // optional
	GasPrice         uint64
	GasBudget        uint64
}

func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	req *CreateAndSendRequestRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	anchorRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.AnchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBCreateAndSendRequest(
		ptb,
		req.PackageID,
		*anchorRef.ObjectID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: req.AssetsBagRef}),
		req.Message,
		req.Allowance,
		req.OnchainGasBudget,
	)
	return c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
}

type CreateAndSendRequestWithAssetsRequest struct {
	Signer           cryptolib.Signer
	PackageID        iotago.PackageID
	AnchorAddress    *iotago.ObjectID
	Assets           *iscmove.Assets
	Message          *iscmove.Message
	Allowance        *iscmove.Assets
	OnchainGasBudget uint64
	GasPayments      []*iotago.ObjectRef // optional
	GasPrice         uint64
	GasBudget        uint64
}

func (c *Client) CreateAndSendRequestWithAssets(
	ctx context.Context,
	req *CreateAndSendRequestWithAssetsRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	anchorRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.AnchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	gotAllCoins, err := c.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{Owner: req.Signer.Address().AsIotaAddress()})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	placedCoins := []lo.Tuple2[*iotajsonrpc.Coin, uint64]{}
	// assume we can find it in the first page
	for cointype, bal := range req.Assets.Coins {
		cumulativedBal := uint64(0)
		tmpPlacedCoins := []lo.Tuple2[*iotajsonrpc.Coin, uint64]{}
		for _, gotCoin := range gotAllCoins.Data {
			if !lo.Must(iotago.IsSameResource(cointype, gotCoin.CoinType)) {
				continue
			}
			if lo.ContainsBy(req.GasPayments, func(ref *iotago.ObjectRef) bool {
				return ref.ObjectID.Equals(*gotCoin.CoinObjectID)
			}) {
				continue
			}
			if bal.Uint64() > cumulativedBal {
				gotCoin := gotCoin
				tmpPlacedCoins = append(tmpPlacedCoins, lo.Tuple2[*iotajsonrpc.Coin, uint64]{A: gotCoin, B: gotCoin.Balance.Uint64()})
				cumulativedBal += gotCoin.Balance.Uint64()
			}
		}
		if bal.Uint64() > cumulativedBal {
			return nil, fmt.Errorf("cannot find coin for type %s", cointype)
		}
		placedCoins = append(placedCoins, tmpPlacedCoins...)
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagNew(ptb, req.PackageID, req.Signer.Address())
	argAssetsBag := ptb.LastCommandResultArg()
	for _, tuple := range placedCoins {
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: tuple.A.Ref()}),
			tuple.B,
			tuple.A.CoinType,
		)
	}
	ptb = PTBCreateAndSendRequest(
		ptb,
		req.PackageID,
		*anchorRef.ObjectID,
		argAssetsBag,
		req.Message,
		req.Allowance,
		req.OnchainGasBudget,
	)
	return c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
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
