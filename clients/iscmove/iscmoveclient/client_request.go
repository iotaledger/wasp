package iscmoveclient

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
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
) (*iotagraphql.IotaTransactionBlockResponse, error) {
	anchorRes, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{ObjectID: req.AnchorAddress})
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

func (c *Client) selectProperGasCoinAndBalance(ctx context.Context, req *CreateAndSendRequestWithAssetsRequest) ([]*iotagraphql.Coin, uint64, error) {
	iotaBalance := req.Assets.BaseToken()

	coinOptions, err := c.GetCoinObjsForTargetAmount(ctx, req.Signer.Address().AsIotaAddress(), iotaBalance.Uint64(), iotagraphql.DefaultGasBudget)
	if err != nil {
		return nil, 0, err
	}

	coin, err := coinOptions.PickMultipleCoinsNoLess(iotaBalance.Uint64())
	if err != nil {
		return nil, 0, err
	}

	return coin, iotaBalance.Uint64(), nil
}

func (c *Client) collectPlacedCoins(
	ctx context.Context,
	req *CreateAndSendRequestWithAssetsRequest,
) ([]lo.Tuple2[*iotagraphql.Coin, uint64], error) {
	var placedCoins []lo.Tuple2[*iotagraphql.Coin, uint64]
	// Query for each specific coin type needed
	for cointype, bal := range req.Assets.Coins.Iterate() {
		if lo.Must(iotago.IsSameResource(cointype.String(), iotagraphql.IotaCoinType.String())) {
			continue
		}

		// Query for this specific coin type
		coinTypeStr := cointype.String()
		coinsOfType, err := c.GetCoins(ctx, iotagraphql.GetCoinsRequest{
			Owner:    req.Signer.Address().AsIotaAddress(),
			CoinType: &coinTypeStr,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get coins of type %s: %w", cointype, err)
		}

		coin, ok := lo.Find(coinsOfType.Data, func(coin *iotagraphql.Coin) bool {
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

		// Update the coin ref to get the latest version
		coinRef := coin.Ref()
		updatedRef, err := c.UpdateObjectRef(ctx, coinRef)
		if err != nil {
			return nil, fmt.Errorf("failed to update coin ref for type %s: %w", cointype, err)
		}
		coin.Version = graphqltypes.NewBigInt(updatedRef.Version)
		coin.Digest = updatedRef.Digest
		// Use the unwrapped coin type from the Assets iterator (e.g. "0x...::testcoin::TESTCOIN")
		// instead of the GraphQL response type which includes the Coin<> wrapper
		coin.CoinType = cointype

		placedCoins = append(placedCoins, lo.Tuple2[*iotagraphql.Coin, uint64]{A: coin, B: bal.Uint64()})
	}
	return placedCoins, nil
}

//nolint:funlen
func (c *Client) CreateAndSendRequestWithAssets(
	ctx context.Context,
	req *CreateAndSendRequestWithAssetsRequest,
) (*iotagraphql.IotaTransactionBlockResponse, error) {
	anchorRes, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{ObjectID: req.AnchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	placedCoins, err := c.collectPlacedCoins(ctx, req)
	if err != nil {
		return nil, err
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBAssetsBagNew(ptb, req.PackageID, req.Signer.Address())
	argAssetsBag := ptb.LastCommandResultArg()

	// Select IOTA coin first
	gasCoins, balance, err := c.selectProperGasCoinAndBalance(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find an IOTA coin with proper balance ref: %w", err)
	}

	if len(gasCoins) > 1 {
		primaryIdx := 0
		primaryBal := gasCoins[0].Balance.Uint64()
		for i := 1; i < len(gasCoins); i++ {
			bal := gasCoins[i].Balance.Uint64()
			if bal > primaryBal {
				primaryIdx = i
				primaryBal = bal
			}
		}
		if primaryIdx != 0 {
			gasCoins[0], gasCoins[primaryIdx] = gasCoins[primaryIdx], gasCoins[0]
		}
	}

	if balance > 0 {
		if gasCoins[0].Balance.Uint64() < balance {
			if len(gasCoins) == 1 {
				return nil, fmt.Errorf("insufficient balance in gas coin: need %d, have %d", balance, gasCoins[0].Balance.Uint64())
			}
			coinsToMerge := make([]iotago.Argument, 0, len(gasCoins)-1)
			for i := 1; i < len(gasCoins); i++ {
				coinsToMerge = append(coinsToMerge, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: gasCoins[i].Ref()}))
			}
			ptb.Command(iotago.Command{
				MergeCoins: &iotago.ProgrammableMergeCoins{
					Destination: iotago.GetArgumentGasCoin(),
					Sources:     coinsToMerge,
				},
			})
		}
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			iotago.GetArgumentGasCoin(),
			iotagraphql.CoinValue(balance),
			iotagraphql.IotaCoinType,
		)
	}

	// Then the rest of the coins
	for _, tuple := range placedCoins {
		ptb = PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: tuple.A.Ref()}),
			iotagraphql.CoinValue(tuple.B),
			tuple.A.CoinType,
		)
	}

	// Place the non-coin objects
	for id, t := range req.Assets.Objects.Iterate() {
		objRes, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{ObjectID: &id})
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
	var gasCoinRefs []*iotago.ObjectRef
	for _, gasCoin := range gasCoins {
		gasCoinRefs = append(gasCoinRefs, gasCoin.Ref())
	}
	return c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		gasCoinRefs,
		req.GasPrice,
		req.GasBudget,
	)
}

func (c *Client) GetRequestFromObjectID(
	ctx context.Context,
	reqID *iotago.ObjectID,
) (*iscmove.RefWithObject[iscmove.Request], error) {
	getObjectResponse, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{
		ObjectID: reqID,
		Options:  &iotagraphql.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get request content: %w", err)
	}
	if getObjectResponse.Data == nil {
		return nil, fmt.Errorf("request %s not found", *reqID)
	}
	return c.parseRequestAndFetchAssetsBag(ctx, getObjectResponse.Data)
}

func (c *Client) parseRequestAndFetchAssetsBag(ctx context.Context, obj *iotagraphql.IotaObjectData) (*iscmove.RefWithObject[iscmove.Request], error) {
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
	err := iotagraphql.UnmarshalBCS(obj.Bcs.MoveObject.BcsBytes, &intermediateRequest)
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

func (c *Client) pullRequests(ctx context.Context, packageID iotago.Address, anchorAddress *iotago.ObjectID, maxAmountOfRequests int) (map[iotago.ObjectID]*iotagraphql.IotaObjectData, error) {
	pulledRequests := make(map[iotago.ObjectID]*iotagraphql.IotaObjectData, maxAmountOfRequests)

	query := &iotagraphql.IotaObjectResponseQuery{
		Filter: &iotagraphql.IotaObjectDataFilter{
			StructType: &iotago.StructTag{
				Address: &packageID,
				Module:  iscmove.RequestModuleName,
				Name:    iscmove.RequestObjectName,
			},
		},
		Options: &iotagraphql.IotaObjectDataOptions{
			ShowBcs:   true,
			ShowOwner: true,
		},
	}

	var cursor *iotago.ObjectID
	for len(pulledRequests) < maxAmountOfRequests {
		objs, err := c.GetOwnedObjects(ctx, iotagraphql.GetOwnedObjectsRequest{
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
