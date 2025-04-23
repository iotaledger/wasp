package iscmoveclient

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type CreateAndSendRequestRequest struct {
	Signer        cryptolib.Signer
	PackageID     iotago.PackageID
	AnchorAddress *iotago.ObjectID
	AssetsBagRef  *iotago.ObjectRef
	Message       *iscmove.Message
	// AllowanceBCS is either empty or a BCS-encoded iscmove.Allowance
	AllowanceBCS     []byte
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
		req.AllowanceBCS,
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
	Signer        cryptolib.Signer
	PackageID     iotago.PackageID
	AnchorAddress *iotago.ObjectID
	Assets        *iscmove.Assets
	Message       *iscmove.Message
	// AllowanceBCS is either empty or a BCS-encoded iscmove.Allowance
	AllowanceBCS     []byte
	OnchainGasBudget uint64
	GasPayments      []*iotago.ObjectRef // optional
	GasPrice         uint64
	GasBudget        uint64
}

func (c *Client) selectProperGasCoinAndBalance(ctx context.Context, req *CreateAndSendRequestWithAssetsRequest) (*iotajsonrpc.Coin, uint64, error) {
	iotaBalance := req.Assets.BaseToken()

	coinOptions, err := c.GetCoinObjsForTargetAmount(ctx, req.Signer.Address().AsIotaAddress(), iotaBalance, iotaclient.DefaultGasBudget)
	if err != nil {
		return nil, 0, err
	}

	coin, err := coinOptions.PickCoinNoLess(iotaBalance)
	if err != nil {
		return nil, 0, err
	}

	return coin, iotaBalance, nil
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

	allCoins, err := c.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{Owner: req.Signer.Address().AsIotaAddress()})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	var placedCoins []lo.Tuple2[*iotajsonrpc.Coin, uint64]
	// assume we can find it in the first page
	for cointype, bal := range req.Assets.Coins.Iterate() {
		if lo.Must(iotago.IsSameResource(cointype.String(), iotajsonrpc.IotaCoinType.String())) {
			continue
		}

		coin, ok := lo.Find(allCoins.Data, func(coin *iotajsonrpc.Coin) bool {
			if !lo.Must(iotago.IsSameResource(cointype.String(), string(coin.CoinType))) {
				return false
			}
			if lo.ContainsBy(req.GasPayments, func(ref *iotago.ObjectRef) bool {
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
	ptb = PTBAssetsBagNew(ptb, req.PackageID, req.Signer.Address())
	argAssetsBag := ptb.LastCommandResultArg()

	// Select IOTA coin first
	gasCoin, balance, err := c.selectProperGasCoinAndBalance(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find an IOTA coin with proper balance ref: %w", err)
	}

	if balance > 0 {
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			iotago.GetArgumentGasCoin(),
			iotajsonrpc.CoinValue(balance),
			iotajsonrpc.IotaCoinType,
		)
	}

	// Then the rest of the coins
	for _, tuple := range placedCoins {
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: tuple.A.Ref()}),
			iotajsonrpc.CoinValue(tuple.B),
			tuple.A.CoinType,
		)
	}

	// Place the non-coin objects
	for id, t := range req.Assets.Objects.Iterate() {
		objRes, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: &id})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", id, err)
		}
		ref := objRes.Data.Ref()
		ptb = PTBAssetsBagPlaceObject(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: &ref}),
			t,
		)
	}

	ptb = PTBCreateAndSendRequest(
		ptb,
		req.PackageID,
		*anchorRef.ObjectID,
		argAssetsBag,
		req.Message,
		req.AllowanceBCS,
		req.OnchainGasBudget,
	)
	return c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		[]*iotago.ObjectRef{gasCoin.Ref()},
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
	return c.parseRequestAndFetchAssetsBag(ctx, getObjectResponse.Data)
}

func (c *Client) parseRequestAndFetchAssetsBag(ctx context.Context, obj *iotajsonrpc.IotaObjectData) (*iscmove.RefWithObject[iscmove.Request], error) {
	// intermediateMoveRequest is used to decode actual requests coming from move.
	// The only difference between this and MoveRequest is the AssetsBag
	// The Balances in AssetsBagWithBalance are unavailable in the bcs encoded Request coming from L1
	// The type will get mapped into an actual MoveRequest after it has been enriched.
	// It decouples the problem that other types which depend on AssetsBagWithBalances can't properly encode Balances
	// as they have to be ignored. Otherwise, the moveRequest decoding will fail.
	type intermediateMoveRequest struct {
		ID        iotago.ObjectID
		Sender    *cryptolib.Address
		AssetsBag iscmove.Referent[iscmove.AssetsBag]
		Message   iscmove.Message
		Allowance []byte
		GasBudget uint64
	}

	var intermediateRequest intermediateMoveRequest
	err := iotaclient.UnmarshalBCS(obj.Bcs.Data.MoveObject.BcsBytes, &intermediateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	bals, err := c.GetAssetsBagWithBalances(ctx, &intermediateRequest.AssetsBag.Value.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AssetsBag of Request: %w", err)
	}

	req := MoveRequest{
		ID:     intermediateRequest.ID,
		Sender: intermediateRequest.Sender,
		AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
			ID:    intermediateRequest.AssetsBag.ID,
			Value: bals,
		},
		Message:   intermediateRequest.Message,
		Allowance: intermediateRequest.Allowance,
		GasBudget: intermediateRequest.GasBudget,
	}

	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: obj.Ref(),
		Object:    req.ToRequest(),
		Owner:     obj.Owner.AddressOwner,
	}, nil
}

func (c *Client) pullRequests(ctx context.Context, packageID iotago.Address, anchorAddress *iotago.ObjectID, maxAmountOfRequests int) (map[iotago.ObjectID]*iotajsonrpc.IotaObjectData, error) {
	pulledRequests := make(map[iotago.ObjectID]*iotajsonrpc.IotaObjectData, maxAmountOfRequests)

	query := &iotajsonrpc.IotaObjectResponseQuery{
		Filter: &iotajsonrpc.IotaObjectDataFilter{
			StructType: &iotago.StructTag{
				Address: &packageID,
				Module:  iscmove.RequestModuleName,
				Name:    iscmove.RequestObjectName,
			},
		},
		Options: &iotajsonrpc.IotaObjectDataOptions{
			ShowBcs:   true,
			ShowOwner: true,
		},
	}

	var cursor *iotago.ObjectID
	for len(pulledRequests) < maxAmountOfRequests {
		objs, err := c.GetOwnedObjects(ctx, iotaclient.GetOwnedObjectsRequest{
			Address: anchorAddress,
			Query:   query,
			Cursor:  cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch requests: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context error while fetching requests: %w", err)
		}

		if objs == nil || len(objs.Data) == 0 {
			break
		}

		// Process the fetched objects
		for _, req := range objs.Data {
			if req.Data == nil || req.Data.ObjectID == nil {
				continue
			}

			pulledRequests[*req.Data.ObjectID] = req.Data
			if len(pulledRequests) >= maxAmountOfRequests {
				break
			}
		}

		// Update cursor for next iteration
		if objs.NextCursor == nil {
			break
		}

		cursor = objs.NextCursor
	}

	return pulledRequests, nil
}

// GetRequestsSorted pulls all requests owned by a certain address, sorts their ids, and returns a certain amount of them.
// This is needed, so the Consensus has the same requests to work with, because GetOwnedObjects would return a random ordered list.
// Additionally, the mempool has a certain limit we don't want to overflow.
// This function will be called periodically by the Chain Manager to pick up the next amount of transactions.
// This ensures that we don't suddenly load a huge amount of Requests into the mempool which fills up at a lower limit.
func (c *Client) GetRequestsSorted(ctx context.Context, packageID iotago.PackageID, anchorAddress *iotago.ObjectID, maxAmountOfRequests int, cb func(error, *iscmove.RefWithObject[iscmove.Request])) error {
	pulledRequests, err := c.pullRequests(ctx, packageID, anchorAddress, maxAmountOfRequests)
	if err != nil {
		return err
	}

	objectKeys := maps.Keys(pulledRequests)
	sort.Slice(objectKeys, func(i, j int) bool {
		return bytes.Compare(objectKeys[i][:], objectKeys[j][:]) < 0
	})

	var sortedRequestIDs []iotago.ObjectID

	if len(objectKeys) >= maxAmountOfRequests {
		sortedRequestIDs = objectKeys[:maxAmountOfRequests]
	} else {
		sortedRequestIDs = objectKeys
	}

	// TODO: Improve loading of the requests by requesting in parallel
	for _, reqID := range sortedRequestIDs {
		ref, err := c.parseRequestAndFetchAssetsBag(ctx, pulledRequests[reqID])
		cb(err, ref)
	}

	return nil
}

func (c *Client) GetRequests(
	ctx context.Context,
	packageID iotago.PackageID,
	anchorAddress *iotago.ObjectID,
	maxAmountOfRequests int,
) (
	[]*iscmove.RefWithObject[iscmove.Request],
	error,
) {
	requests, err := c.pullRequests(ctx, packageID, anchorAddress, maxAmountOfRequests)
	if err != nil {
		return nil, err
	}

	parsedRequests := make([]*iscmove.RefWithObject[iscmove.Request], 0)

	for _, reqData := range requests {
		req, err := c.parseRequestAndFetchAssetsBag(ctx, reqData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode request: %w", err)
		}

		parsedRequests = append(parsedRequests, req)
	}

	return parsedRequests, nil
}
