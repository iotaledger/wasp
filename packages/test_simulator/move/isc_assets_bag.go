package move

import (
	"encoding/json"
	"fmt"

	"fortio.org/safecast"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var AssetsBagHandlers = map[string]l1.MoveCallFunc{
	"new":                   assetsBagNew,
	"destroy_empty":         assetsBagDestroyEmpty,
	"get_size":              assetsBagGetSize,
	"place_coin":            assetsBagPlaceCoin,
	"place_coin_balance":    assetsBagPlaceCoinBalance,
	"place_asset":           assetsBagPlaceAsset,
	"take_coin_balance":     assetsBagTakeCoinBalance,
	"take_all_coin_balance": assetsBagTakeAllCoinBalance,
	"take_asset":            assetsBagTakeAsset,
}

type AssetsBagValue struct {
	ID   iotago.ObjectID
	Size uint64
}

func assetsBagNew(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, _ []l1.Value) ([]l1.Value, error) {
	id := ctx.FreshID()
	bag := &AssetsBagValue{ID: id, Size: 0}
	typeName := l1.ISCTypeString(ctx.PackageID, "assets_bag", "AssetsBag")
	data := bcs.MustMarshal(bag)
	ctx.Store.Put(&l1.SimObject{
		ID:         id,
		Version:    0,
		Digest:     l1.ComputeDigest(data),
		Owner:      l1.SimOwner{AddressOwner: &ctx.Sender},
		Type:       typeName,
		Data:       data,
		PreviousTx: ctx.TxDigest,
	})
	return []l1.Value{{ObjectID: &id, Raw: bag, Type: typeName}}, nil
}

// destroy_empty(bag)
func assetsBagDestroyEmpty(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("assets_bag::destroy_empty requires 1 argument")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::destroy_empty: %w", err)
	}
	if bag.Size != 0 {
		return nil, fmt.Errorf("assets_bag::destroy_empty: bag is not empty (size=%d)", bag.Size)
	}
	return nil, nil
}

// get_size(bag) -> u64
func assetsBagGetSize(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("assets_bag::get_size requires 1 argument")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::get_size: %w", err)
	}
	return []l1.Value{{Raw: bag.Size, Type: "u64"}}, nil
}

// place_coin<T>(bag, coin)
func assetsBagPlaceCoin(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assets_bag::place_coin requires 2 arguments")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::place_coin: %w", err)
	}

	coinType := firstTypeArg(call)

	var balance uint64
	if args[1].ObjectID != nil {
		obj, ok := ctx.Store.Get(*args[1].ObjectID)
		if !ok {
			return nil, fmt.Errorf("assets_bag::place_coin: coin object not found")
		}
		balance = l1.DecodeCoinObjectBalance(obj.Data)
		ctx.Store.Delete(*args[1].ObjectID)
	} else if bal, ok := args[1].Raw.(*l1.BalanceValue); ok {
		balance = bal.Amount
	} else {
		return nil, fmt.Errorf("assets_bag::place_coin: cannot extract balance from argument")
	}

	placeCoinBalanceInternal(ctx, bag, coinType, balance)
	return nil, nil
}

// place_coin_balance<T>(bag, balance)
func assetsBagPlaceCoinBalance(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assets_bag::place_coin_balance requires 2 arguments")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::place_coin_balance: %w", err)
	}

	coinType := firstTypeArg(call)

	bal, ok := args[1].Raw.(*l1.BalanceValue)
	if !ok {
		return nil, fmt.Errorf("assets_bag::place_coin_balance: argument is not a BalanceValue")
	}

	placeCoinBalanceInternal(ctx, bag, coinType, bal.Amount)
	return nil, nil
}

// place_asset<T>(bag, asset)
func assetsBagPlaceAsset(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assets_bag::place_asset requires 2 arguments")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::place_asset: %w", err)
	}

	assetID := args[1].ObjectID
	if assetID == nil {
		return nil, fmt.Errorf("assets_bag::place_asset: asset has no object ID")
	}

	nameJSON, _ := json.Marshal(assetID.String())
	ctx.Store.AddDynamicField(l1.DynamicField{
		ParentID:   bag.ID,
		Name:       l1.DynFieldName{TypeRepr: l1.ObjectIDTypeString(), JSON: nameJSON},
		ValueObjID: *assetID,
	})
	bag.Size++
	persistBag(ctx, bag)
	return nil, nil
}

// take_coin_balance<T>(bag, amount) -> Balance<T>
func assetsBagTakeCoinBalance(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assets_bag::take_coin_balance requires 2 arguments")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_coin_balance: %w", err)
	}

	coinType := firstTypeArg(call)
	amount, err := extractUint64(args[1])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_coin_balance: %w", err)
	}

	bal, err := takeCoinBalanceInternal(ctx, bag, coinType, amount)
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_coin_balance: %w", err)
	}

	return []l1.Value{{Raw: bal, Type: l1.BalanceTypeString(coinType)}}, nil
}

// take_all_coin_balance<T>(bag) -> Balance<T>
func assetsBagTakeAllCoinBalance(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("assets_bag::take_all_coin_balance requires 1 argument")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_all_coin_balance: %w", err)
	}

	coinType := firstTypeArg(call)
	nameJSON := coinTypeToNameJSON(coinType)

	df, ok := ctx.Store.GetDynamicField(bag.ID, l1.ASCIIStringTypeString(), string(nameJSON))
	if !ok {
		return nil, fmt.Errorf("assets_bag::take_all_coin_balance: no balance for coin type %s", coinType)
	}

	balObj, ok := ctx.Store.Get(df.ValueObjID)
	if !ok {
		return nil, fmt.Errorf("assets_bag::take_all_coin_balance: balance object not found")
	}
	amount := l1.DecodeBalanceValue(balObj.Data)

	ctx.Store.RemoveDynamicField(bag.ID, l1.ASCIIStringTypeString(), string(nameJSON))
	ctx.Store.Delete(df.ValueObjID)
	bag.Size--
	persistBag(ctx, bag)

	return []l1.Value{{Raw: &l1.BalanceValue{CoinType: coinType, Amount: amount}, Type: l1.BalanceTypeString(coinType)}}, nil
}

// take_asset<T>(bag, id) -> T
func assetsBagTakeAsset(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assets_bag::take_asset requires 2 arguments")
	}
	bag, err := extractBag(args[0])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_asset: %w", err)
	}

	assetID, err := extractObjectID(args[1])
	if err != nil {
		return nil, fmt.Errorf("assets_bag::take_asset: %w", err)
	}

	nameJSON, _ := json.Marshal(assetID.String())
	df, ok := ctx.Store.RemoveDynamicField(bag.ID, l1.ObjectIDTypeString(), string(nameJSON))
	if !ok {
		return nil, fmt.Errorf("assets_bag::take_asset: asset not found in bag")
	}

	bag.Size--
	persistBag(ctx, bag)
	return []l1.Value{{ObjectID: &df.ValueObjID, Type: ""}}, nil
}

func placeCoinBalanceInternal(ctx *l1.CallContext, bag *AssetsBagValue, coinType string, amount uint64) {
	nameJSON := coinTypeToNameJSON(coinType)

	df, exists := ctx.Store.GetDynamicField(bag.ID, l1.ASCIIStringTypeString(), string(nameJSON))
	if exists {
		balObj, ok := ctx.Store.Get(df.ValueObjID)
		if ok {
			newBalance := l1.DecodeBalanceValue(balObj.Data) + amount
			balObj.Data = bcs.MustMarshal(&newBalance)
			balObj.Digest = l1.ComputeDigest(balObj.Data)
			ctx.Store.Put(balObj)
		}
	} else {
		balID := ctx.FreshID()
		balData := bcs.MustMarshal(&amount)
		parentID := bag.ID
		balObj := &l1.SimObject{
			ID:         balID,
			Version:    0,
			Digest:     l1.ComputeDigest(balData),
			Owner:      l1.SimOwner{ObjectOwner: &parentID},
			Type:       l1.BalanceTypeString(coinType),
			Data:       balData,
			PreviousTx: ctx.TxDigest,
		}
		ctx.Store.Put(balObj)
		ctx.Store.AddDynamicField(l1.DynamicField{
			ParentID:   bag.ID,
			Name:       l1.DynFieldName{TypeRepr: l1.ASCIIStringTypeString(), JSON: nameJSON},
			ValueObjID: balID,
		})
		bag.Size++
	}
	persistBag(ctx, bag)
}

func takeCoinBalanceInternal(ctx *l1.CallContext, bag *AssetsBagValue, coinType string, amount uint64) (*l1.BalanceValue, error) {
	nameJSON := coinTypeToNameJSON(coinType)

	df, ok := ctx.Store.GetDynamicField(bag.ID, l1.ASCIIStringTypeString(), string(nameJSON))
	if !ok {
		return nil, fmt.Errorf("no balance for coin type %s", coinType)
	}

	balObj, ok := ctx.Store.Get(df.ValueObjID)
	if !ok {
		return nil, fmt.Errorf("balance object not found")
	}

	existing := l1.DecodeBalanceValue(balObj.Data)
	if existing < amount {
		return nil, fmt.Errorf("insufficient balance: have %d, want %d", existing, amount)
	}

	remaining := existing - amount
	if remaining == 0 {
		ctx.Store.RemoveDynamicField(bag.ID, l1.ASCIIStringTypeString(), string(nameJSON))
		ctx.Store.Delete(df.ValueObjID)
		bag.Size--
		persistBag(ctx, bag)
	} else {
		balObj.Data = bcs.MustMarshal(&remaining)
		balObj.Digest = l1.ComputeDigest(balObj.Data)
		ctx.Store.Put(balObj)
	}

	return &l1.BalanceValue{CoinType: coinType, Amount: amount}, nil
}

// coinTypeToNameJSON converts a coin type string to the JSON format used as a dynamic field name.
func coinTypeToNameJSON(coinType string) json.RawMessage {
	name := coinType
	if len(name) > 2 && name[:2] == "0x" {
		name = name[2:]
	}
	b, _ := json.Marshal(name)
	return b
}

func persistBag(ctx *l1.CallContext, bag *AssetsBagValue) {
	ab := iscmove.AssetsBag{ID: bag.ID, Size: bag.Size}
	data := bcs.MustMarshal(&ab)
	if obj, ok := ctx.Store.Get(bag.ID); ok {
		obj.Data = data
		obj.Digest = l1.ComputeDigest(data)
		ctx.Store.Put(obj)
	}
}

func extractBag(v l1.Value) (*AssetsBagValue, error) {
	if bag, ok := v.Raw.(*AssetsBagValue); ok {
		return bag, nil
	}
	if obj, ok := v.Raw.(*l1.SimObject); ok {
		ab, err := bcs.Unmarshal[iscmove.AssetsBag](obj.Data)
		if err != nil {
			return nil, fmt.Errorf("BCS unmarshal AssetsBag failed: %w", err)
		}
		return &AssetsBagValue{ID: ab.ID, Size: ab.Size}, nil
	}
	return nil, fmt.Errorf("expected AssetsBagValue, got %T", v.Raw)
}

func extractUint64(v l1.Value) (uint64, error) {
	switch val := v.Raw.(type) {
	case uint64:
		return val, nil
	case *uint64:
		return *val, nil
	case int64:
		return safecast.Convert[uint64](val)
	case int:
		return safecast.Convert[uint64](val)
	case []byte:
		result, err := bcs.Unmarshal[uint64](val)
		if err != nil {
			return 0, fmt.Errorf("BCS decode u64 failed: %w", err)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("expected uint64, got %T", v.Raw)
	}
}

func extractObjectID(v l1.Value) (*iotago.ObjectID, error) {
	if v.ObjectID != nil {
		return v.ObjectID, nil
	}
	if raw, ok := v.Raw.([]byte); ok && len(raw) == 32 {
		var id iotago.ObjectID
		copy(id[:], raw)
		return &id, nil
	}
	return nil, fmt.Errorf("expected ObjectID, got %T (ObjectID field is nil)", v.Raw)
}

func firstTypeArg(call *iotago.ProgrammableMoveCall) string {
	if len(call.TypeArguments) > 0 {
		return call.TypeArguments[0].String()
	}
	return ""
}
