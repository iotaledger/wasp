package move

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var AnchorHandlers = map[string]l1.MoveCallFunc{
	"start_new_chain":                   anchorStartNewChain,
	"borrow_assets":                     anchorBorrowAssets,
	"return_assets_from_borrow":         anchorReturnAssetsFromBorrow,
	"receive_request":                   anchorReceiveRequest,
	"transition":                        anchorTransition,
	"create_anchor_with_assets_bag_ref": anchorCreateWithAssetsBagRef,
	"update_anchor_state_for_migration": anchorUpdateStateForMigration,
	"destroy":                           anchorDestroy,
	"place_coin_for_migration":          anchorPlaceCoinForMigration,
	"place_coin_balance_for_migration":  anchorPlaceCoinBalanceForMigration,
	"place_asset_for_migration":         anchorPlaceAssetForMigration,
}

type AnchorValue struct {
	ID            iotago.ObjectID
	Assets        *ReferentValue // contains AssetsBagValue
	StateMetadata []byte
	StateIndex    uint32
}

// start_new_chain(state_metadata, opt_coin, ctx) -> Anchor
func anchorStartNewChain(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("anchor::start_new_chain requires 2 arguments")
	}

	stateMetadata, err := extractBytes(args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::start_new_chain: state_metadata: %w", err)
	}

	optCoin := args[1].Raw
	bagID := ctx.FreshID()
	bag := &AssetsBagValue{ID: bagID, Size: 0}

	if opt, ok := optCoin.(*OptionValue); ok && opt.IsSome && opt.Value != nil {
		coinVal := *opt.Value
		var balance uint64
		if coinVal.ObjectID != nil {
			obj, ok := ctx.Store.Get(*coinVal.ObjectID)
			if !ok {
				return nil, fmt.Errorf("anchor::start_new_chain: coin object not found")
			}
			balance = l1.DecodeCoinObjectBalance(obj.Data)
			ctx.Store.Delete(*coinVal.ObjectID)
		} else if bal, ok := coinVal.Raw.(*l1.BalanceValue); ok {
			balance = bal.Amount
		}
		if balance > 0 {
			iotaCoinType := l1.IotaCoinTypeStr
			placeCoinBalanceInternal(ctx, bag, iotaCoinType, balance)
		}
	}

	referentID := ctx.FreshID()
	bagValue := l1.Value{ObjectID: &bagID, Raw: bag, Type: "AssetsBag"}
	referent := &ReferentValue{
		ID:    referentID,
		Value: &bagValue,
	}

	anchorID := ctx.FreshID()
	anchor := &AnchorValue{
		ID:            anchorID,
		Assets:        referent,
		StateMetadata: stateMetadata,
		StateIndex:    0,
	}

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return []l1.Value{{ObjectID: &anchorID, Raw: anchor, Type: l1.ISCTypeString(ctx.PackageID, iscmove.AnchorModuleName, iscmove.AnchorObjectName)}}, nil
}

// borrow_assets(anchor) -> (AssetsBag, Borrow)
func anchorBorrowAssets(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("anchor::borrow_assets requires 1 argument")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::borrow_assets: %w", err)
	}

	if anchor.Assets.Value == nil {
		return nil, fmt.Errorf("anchor::borrow_assets: assets already borrowed")
	}

	inner := *anchor.Assets.Value
	token := &l1.BorrowToken{
		ReferentID: anchor.Assets.ID,
		RefAddr:    anchor.Assets.ID,
	}
	anchor.Assets.Value = nil

	return []l1.Value{inner, {Raw: token, Type: "Borrow"}}, nil
}

// return_assets_from_borrow(anchor, assets_bag, borrow)
func anchorReturnAssetsFromBorrow(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("anchor::return_assets_from_borrow requires 3 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::return_assets_from_borrow: %w", err)
	}

	token, ok := args[2].Raw.(*l1.BorrowToken)
	if !ok {
		return nil, fmt.Errorf("anchor::return_assets_from_borrow: third argument is not a BorrowToken")
	}
	if token.ReferentID != anchor.Assets.ID {
		return nil, fmt.Errorf("anchor::return_assets_from_borrow: borrow token does not match referent")
	}

	anchor.Assets.Value = &args[1]

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

// receive_request(anchor, receiving) -> (Receipt, AssetsBag)
func anchorReceiveRequest(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("anchor::receive_request requires 2 arguments")
	}

	reqID := args[1].ObjectID
	if reqID == nil {
		return nil, fmt.Errorf("anchor::receive_request: receiving argument has no object ID")
	}

	reqObj, ok := ctx.Store.Get(*reqID)
	if !ok {
		return nil, fmt.Errorf("anchor::receive_request: request %s not found", reqID.String())
	}

	destroyResult, err := requestDestroy(ctx, nil, []l1.Value{{ObjectID: reqID, Raw: reqObj, Type: reqObj.Type}})
	if err != nil {
		return nil, fmt.Errorf("anchor::receive_request: destroy: %w", err)
	}

	receipt := &iscmove.Receipt{RequestID: *reqID}

	return []l1.Value{
		{Raw: receipt, Type: "Receipt"},
		destroyResult[1],
	}, nil
}

// transition(anchor, new_state_metadata, receipts)
func anchorTransition(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("anchor::transition requires 3 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::transition: %w", err)
	}

	newMetadata, err := extractBytes(args[1])
	if err != nil {
		return nil, fmt.Errorf("anchor::transition: new_state_metadata: %w", err)
	}

	anchor.StateMetadata = newMetadata
	anchor.StateIndex++

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

// create_anchor_with_assets_bag_ref(assets_bag, ctx) -> Anchor
func anchorCreateWithAssetsBagRef(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("anchor::create_anchor_with_assets_bag_ref requires 1 argument")
	}

	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::create_anchor_with_assets_bag_ref: %w", err)
	}

	referentID := ctx.FreshID()
	bagValue := l1.Value{ObjectID: &bag.ID, Raw: bag, Type: "AssetsBag"}
	referent := &ReferentValue{ID: referentID, Value: &bagValue}

	anchorID := ctx.FreshID()
	anchor := &AnchorValue{
		ID:            anchorID,
		Assets:        referent,
		StateMetadata: nil,
		StateIndex:    0,
	}

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return []l1.Value{{ObjectID: &anchorID, Raw: anchor, Type: l1.ISCTypeString(ctx.PackageID, iscmove.AnchorModuleName, iscmove.AnchorObjectName)}}, nil
}

// update_anchor_state_for_migration(anchor, state_metadata, state_index)
func anchorUpdateStateForMigration(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("anchor::update_anchor_state_for_migration requires 3 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::update_anchor_state_for_migration: %w", err)
	}

	metadata, err := extractBytes(args[1])
	if err != nil {
		return nil, fmt.Errorf("anchor::update_anchor_state_for_migration: state_metadata: %w", err)
	}

	stateIndex, err := extractUint32(args[2])
	if err != nil {
		return nil, fmt.Errorf("anchor::update_anchor_state_for_migration: state_index: %w", err)
	}

	anchor.StateMetadata = metadata
	anchor.StateIndex = stateIndex

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

// destroy(anchor) -> AssetsBag
func anchorDestroy(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("anchor::destroy requires 1 argument")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("anchor::destroy: %w", err)
	}

	if anchor.Assets.Value == nil {
		return nil, fmt.Errorf("anchor::destroy: assets still borrowed")
	}

	bagValue := *anchor.Assets.Value
	ctx.Store.Delete(anchor.ID)

	return []l1.Value{bagValue}, nil
}

// place_coin_for_migration<T>(anchor, coin)
func anchorPlaceCoinForMigration(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("anchor::place_coin_for_migration requires 2 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, err
	}

	if anchor.Assets.Value == nil {
		return nil, fmt.Errorf("anchor::place_coin_for_migration: assets still borrowed")
	}
	bag, err := extractBag(*anchor.Assets.Value)
	if err != nil {
		return nil, err
	}

	coinType := firstTypeArg(call)
	var balance uint64
	if args[1].ObjectID != nil {
		obj, ok := ctx.Store.Get(*args[1].ObjectID)
		if !ok {
			return nil, fmt.Errorf("coin not found")
		}
		balance = l1.DecodeCoinObjectBalance(obj.Data)
		ctx.Store.Delete(*args[1].ObjectID)
	}
	if balance > 0 {
		placeCoinBalanceInternal(ctx, bag, coinType, balance)
	}

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

// place_coin_balance_for_migration<T>(anchor, balance)
func anchorPlaceCoinBalanceForMigration(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("anchor::place_coin_balance_for_migration requires 2 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, err
	}

	if anchor.Assets.Value == nil {
		return nil, fmt.Errorf("anchor::place_coin_balance_for_migration: assets still borrowed")
	}
	bag, err := extractBag(*anchor.Assets.Value)
	if err != nil {
		return nil, err
	}

	coinType := firstTypeArg(call)
	bal, ok := args[1].Raw.(*l1.BalanceValue)
	if !ok {
		return nil, fmt.Errorf("expected BalanceValue")
	}
	placeCoinBalanceInternal(ctx, bag, coinType, bal.Amount)

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

// place_asset_for_migration<T>(anchor, asset)
func anchorPlaceAssetForMigration(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("anchor::place_asset_for_migration requires 2 arguments")
	}

	anchor, err := resolveAnchor(ctx, args[0])
	if err != nil {
		return nil, err
	}

	if anchor.Assets.Value == nil {
		return nil, fmt.Errorf("anchor::place_asset_for_migration: assets still borrowed")
	}

	_, err = assetsBagPlaceAsset(ctx, nil, []l1.Value{*anchor.Assets.Value, args[1]})
	if err != nil {
		return nil, err
	}

	anchorObj := anchorToSimObject(ctx, anchor)
	ctx.Store.Put(anchorObj)

	return nil, nil
}

func resolveAnchor(ctx *l1.CallContext, v l1.Value) (*AnchorValue, error) {
	if anchor, ok := v.Raw.(*AnchorValue); ok {
		return anchor, nil
	}

	if v.ObjectID != nil {
		obj, ok := ctx.Store.Get(*v.ObjectID)
		if !ok {
			return nil, fmt.Errorf("anchor object %s not found", v.ObjectID.String())
		}
		return anchorFromSimObject(obj)
	}

	return nil, fmt.Errorf("expected AnchorValue, got %T", v.Raw)
}

func anchorFromSimObject(obj *l1.SimObject) (*AnchorValue, error) {
	moveAnchor, err := bcs.Unmarshal[iscmove.Anchor](obj.Data)
	if err != nil {
		return nil, fmt.Errorf("BCS unmarshal anchor failed: %w", err)
	}

	bagID := moveAnchor.Assets.ID
	var bagValue *l1.Value
	if moveAnchor.Assets.Value != nil {
		v := l1.Value{
			ObjectID: &moveAnchor.Assets.Value.ID,
			Raw: &AssetsBagValue{
				ID:   moveAnchor.Assets.Value.ID,
				Size: moveAnchor.Assets.Value.Size,
			},
			Type: "AssetsBag",
		}
		bagValue = &v
	}

	anchor := &AnchorValue{
		ID: obj.ID,
		Assets: &ReferentValue{
			ID:    bagID,
			Value: bagValue,
		},
		StateMetadata: moveAnchor.StateMetadata,
		StateIndex:    moveAnchor.StateIndex,
	}

	return anchor, nil
}

func anchorToSimObject(ctx *l1.CallContext, anchor *AnchorValue) *l1.SimObject {
	var assetsBagValue *iscmove.AssetsBag
	if anchor.Assets.Value != nil {
		if bag, ok := anchor.Assets.Value.Raw.(*AssetsBagValue); ok {
			assetsBagValue = &iscmove.AssetsBag{
				ID:   bag.ID,
				Size: bag.Size,
			}
		}
	}

	moveAnchor := iscmove.Anchor{
		ID: anchor.ID,
		Assets: iscmove.Referent[iscmove.AssetsBag]{
			ID:    anchor.Assets.ID,
			Value: assetsBagValue,
		},
		StateMetadata: anchor.StateMetadata,
		StateIndex:    anchor.StateIndex,
	}

	data, err := bcs.Marshal(&moveAnchor)
	if err != nil {
		panic(fmt.Sprintf("BCS marshal anchor failed: %v", err))
	}

	sender := ctx.Sender
	return &l1.SimObject{
		ID:         anchor.ID,
		Version:    0,
		Digest:     l1.ComputeDigest(data),
		Owner:      l1.SimOwner{AddressOwner: &sender},
		Type:       l1.ISCTypeString(ctx.PackageID, iscmove.AnchorModuleName, iscmove.AnchorObjectName),
		Data:       data,
		PreviousTx: ctx.TxDigest,
	}
}
