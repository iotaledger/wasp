package clients

import (
	"context"
	"fmt"
	"os"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type BindingClientL2 struct {
	RpcURL   string
	l1client *BindingClient
}

func NewBindingClientL2(rpcUrl string, l1Client *BindingClient) *BindingClientL2 {
	var client BindingClientL2
	client.l1client = l1Client

	return &client
}

func (c *BindingClientL2) StartNewChain(
	ctx context.Context,
	req *iscmoveclient.StartNewChainRequest,
) (*iscmove.AnchorWithRef, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argInitCoin iotago.Argument
	if req.InitCoinRef != nil {
		ptb = PTBOptionSomeIotaCoin(ptb, req.InitCoinRef)
	} else {
		ptb = PTBOptionNoneIotaCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = iscmoveclient.PTBStartNewChain(ptb, req.PackageID, req.StateMetadata, argInitCoin, req.AnchorOwner)
	txnResponse, err := c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
	if err != nil {
		return nil, fmt.Errorf("start new chain PTB failed: %w", err)
	}

	anchorRef, err := txnResponse.GetCreatedObjectByName(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to GetCreatedObjectInfo: %w", err)
	}
	return c.GetAnchorFromObjectID(ctx, anchorRef.ObjectID)
}
func (c *BindingClientL2) UpdateAnchorStateMetadata(ctx context.Context, req *iscmoveclient.UpdateAnchorStateMetadataRequest) (bool, error) {
	panic("implement me")
}
func (c *BindingClientL2) CreateAndSendRequest(
	ctx context.Context,
	req *iscmoveclient.CreateAndSendRequestRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	panic("implement me")
}
func (c *BindingClientL2) ReceiveRequestsAndTransition(
	ctx context.Context,
	req *iscmoveclient.ReceiveRequestsAndTransitionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	panic("implement me")
}
func (c *BindingClientL2) GetAssetsBagWithBalances(
	ctx context.Context,
	assetsBagID *iotago.ObjectID,
) (*iscmove.AssetsBagWithBalances, error) {
	panic("implement me")
}
func (c *BindingClientL2) CreateAndSendRequestWithAssets(
	ctx context.Context,
	req *iscmoveclient.CreateAndSendRequestWithAssetsRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	anchorRes, err := c.l1client.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.AnchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()

	allCoins, err := c.l1client.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{Owner: req.Signer.Address().AsIotaAddress()})
	if err != nil {
		return nil, fmt.Errorf("failed to get all coins: %w", err)
	}
	var placedCoins []struct {
		coin   *iotajsonrpc.Coin
		amount uint64
	}
	// assume we can find it in the first page
	for cointype, bal := range req.Assets.Coins.Iterate() {
		same, _ := iotago.IsSameResource(cointype.String(), iotajsonrpc.IotaCoinType.String())
		if same {
			continue
		}

		coin, ok := func() (*iotajsonrpc.Coin, bool) {
			for _, c := range allCoins.Data {
				same, _ := iotago.IsSameResource(cointype.String(), string(c.CoinType))
				if !same {
					continue
				}
				// Check if this coin is already used in gas payments
				isGasPayment := false
				for _, ref := range req.GasPayments {
					if ref.ObjectID.Equals(*c.CoinObjectID) {
						isGasPayment = true
						break
					}
				}
				if isGasPayment {
					continue
				}
				if c.Balance.Uint64() >= bal.Uint64() {
					return c, true
				}
			}
			return nil, false
		}()
		if !ok {
			return nil, fmt.Errorf("cannot find coin for type %s", cointype)
		}
		placedCoins = append(placedCoins, struct {
			coin   *iotajsonrpc.Coin
			amount uint64
		}{coin: coin, amount: bal.Uint64()})
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = iscmoveclient.PTBAssetsBagNew(ptb, req.PackageID, req.Signer.Address())
	argAssetsBag := ptb.LastCommandResultArg()

	// Select IOTA coin first
	iotaBalance := req.Assets.BaseToken()
	gasCoins, err := c.l1client.GetCoinObjsForTargetAmount(ctx, req.Signer.Address().AsIotaAddress(), iotaBalance.Uint64(), iotaclient.DefaultGasBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find an IOTA coin with proper balance ref: %w", err)
	}

	balance := iotaBalance.Uint64()
	if balance > 0 {
		ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
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
		ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: tuple.coin.Ref()}),
			iotajsonrpc.CoinValue(tuple.amount),
			tuple.coin.CoinType,
		)
	}

	// Place the non-coin objects
	for id, t := range req.Assets.Objects.Iterate() {
		objRes, err := c.l1client.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: &id})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", id, err)
		}
		ref := objRes.Data.Ref()
		ptb = iscmoveclient.PTBAssetsBagPlaceObject(
			ptb,
			req.PackageID,
			argAssetsBag,
			ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: &ref}),
			t,
		)
	}

	ptb = iscmoveclient.PTBCreateAndSendRequest(
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
func (c *BindingClientL2) GetAnchorFromObjectID(
	ctx context.Context,
	anchorObjectID *iotago.ObjectID,
) (*iscmove.RefWithObject[iscmove.Anchor], error) {
	getObjectResponse, err := c.l1client.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	if getObjectResponse.Error != nil {
		return nil, fmt.Errorf("failed to get anchor content: %s", getObjectResponse.Error.Data.String())
	}
	return decodeAnchorBCS(
		getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes,
		getObjectResponse.Data.Ref(),
		getObjectResponse.Data.Owner.AddressOwner,
	)
}
func (c *BindingClientL2) GetRequestFromObjectID(
	ctx context.Context,
	reqID *iotago.ObjectID,
) (*iscmove.RefWithObject[iscmove.Request], error) {
	panic("implement me")
}
func (c *BindingClientL2) GetCoin(
	ctx context.Context,
	coinID *iotago.ObjectID,
) (*iscmoveclient.MoveCoin, error) {
	panic("implement me")
}
func PTBOptionSome(
	ptb *iotago.ProgrammableTransactionBuilder,
	objTypeTag iotago.TypeTag,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDMoveStdlib,
				Module:        "option",
				Function:      "some",
				TypeArguments: []iotago.TypeTag{objTypeTag},
				Arguments: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: objRef}),
				},
			},
		},
	)
	return ptb
}

func PTBOptionSomeIotaCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	return PTBOptionSome(ptb, *iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iota::IOTA>"), objRef)
}

func PTBOptionNoneIotaCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDMoveStdlib,
				Module:        "option",
				Function:      "none",
				TypeArguments: []iotago.TypeTag{*iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iota::IOTA>")},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	return ptb
}

func (c *BindingClientL2) SignAndExecutePTB(
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
		gasPayments, err = c.l1client.FindCoinsForGasPayment(
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

	if os.Getenv("DEBUG") != "" {
		pt.Print("-- SignAndExecutePTB -- ")
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
	txnResponse, err := c.l1client.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:        true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
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

func decodeAnchorBCS(bcsBytes iotago.Base64Data, ref iotago.ObjectRef, owner *iotago.Address) (*iscmove.AnchorWithRef, error) {
	var moveAnchor iscmove.Anchor
	err := iotaclient.UnmarshalBCS(bcsBytes, &moveAnchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &iscmove.AnchorWithRef{
		ObjectRef: ref,
		Object:    &moveAnchor,
		Owner:     owner,
	}, nil
}
