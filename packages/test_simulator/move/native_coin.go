package move

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var CoinHandlers = map[string]l1.MoveCallFunc{
	"from_balance": coinFromBalance,
	"into_balance": coinIntoBalance,
	"value":        coinValue,
}

// coin::from_balance<T>(balance: Balance<T>, ctx) -> Coin<T>
func coinFromBalance(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("coin::from_balance requires 1 argument")
	}
	bal, ok := args[0].Raw.(*l1.BalanceValue)
	if !ok {
		return nil, fmt.Errorf("coin::from_balance: argument is not a BalanceValue")
	}

	coinType := bal.CoinType
	if len(call.TypeArguments) > 0 {
		coinType = call.TypeArguments[0].String()
	}

	coinID := ctx.FreshID()
	coinObj := createCoinObject(ctx, coinID, coinType, bal.Amount)
	ctx.Store.Put(coinObj)

	return []l1.Value{{ObjectID: &coinID, Type: l1.CoinTypeString(coinType)}}, nil
}

// coin::into_balance<T>(coin: Coin<T>) -> Balance<T>
func coinIntoBalance(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("coin::into_balance requires 1 argument")
	}

	coinType := ""
	if len(call.TypeArguments) > 0 {
		coinType = call.TypeArguments[0].String()
	}

	if args[0].ObjectID != nil {
		obj, ok := ctx.Store.Get(*args[0].ObjectID)
		if !ok {
			return nil, fmt.Errorf("coin::into_balance: coin object not found")
		}
		balance := l1.DecodeCoinObjectBalance(obj.Data)
		if coinType == "" {
			if rt, err := iotago.NewResourceType(obj.Type); err == nil && rt.SubType1 != nil {
				coinType = rt.SubType1.String()
			}
		}
		ctx.Store.Delete(*args[0].ObjectID)

		return []l1.Value{{Raw: &l1.BalanceValue{CoinType: coinType, Amount: balance}, Type: l1.BalanceTypeString(coinType)}}, nil
	}

	return nil, fmt.Errorf("coin::into_balance: argument has no object ID")
}

// coin::value<T>(coin: &Coin<T>) -> u64
func coinValue(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("coin::value requires 1 argument")
	}
	if bal, ok := args[0].Raw.(*l1.BalanceValue); ok {
		return []l1.Value{{Raw: bal.Amount, Type: "u64"}}, nil
	}
	return nil, fmt.Errorf("coin::value: unsupported argument type")
}

func createCoinObject(ctx *l1.CallContext, id iotago.ObjectID, coinType string, balance uint64) *l1.SimObject {
	data := encodeCoinForBCS(id, balance)
	addr := ctx.Sender
	return &l1.SimObject{
		ID:         id,
		Version:    0,
		Digest:     l1.ComputeDigest(data),
		Owner:      l1.SimOwner{AddressOwner: &addr},
		Type:       l1.CoinTypeString(coinType),
		Data:       data,
		PreviousTx: ctx.TxDigest,
	}
}

func encodeCoinForBCS(id iotago.ObjectID, balance uint64) []byte {
	return bcs.MustMarshal(&iscmoveclient.MoveCoin{ID: id, Balance: balance})
}
