package iotagraphql

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

type (
	ExecuteTransactionBlockResponse  = graphqltypes.ExecuteTransactionBlockResponse
	TryGetPastObjectResponse         = graphqltypes.TryGetPastObjectResponse
	BigInt                           = graphqltypes.BigInt
	CoinType                         = graphqltypes.CoinType
	Coins                            = graphqltypes.Coins
	GetCoinsResponse                 = graphqltypes.GetCoinsResponse
	Balance                          = graphqltypes.Balance
	TransactionBytes                 = graphqltypes.TransactionBytes
	CoinValue                        = graphqltypes.CoinValue
	Coin                             = graphqltypes.Coin
	GetDynamicFieldObjectResponse    = graphqltypes.GetDynamicFieldObjectResponse
	GetLatestIotaSystemStateResponse = graphqltypes.GetLatestIotaSystemStateResponse
)

type IotaEvent struct {
	PackageID         *iotago.ObjectID
	TransactionModule string
	Sender            *iotago.Address
	Bcs               []byte
}

type IotaEventFilter struct {
	Package        *iotago.ObjectID
	MoveModule     *IotaEventFilterMoveModule
	MoveEventType  *iotago.StructTag
	MoveEventField *IotaEventFilterMoveEventField
	Sender         *iotago.Address
	And            *IotaAndOrEventFilter
}

type IotaEventFilterMoveModule struct {
	Package *iotago.ObjectID
	Module  string
}

type IotaEventFilterMoveEventField struct {
	Path  string
	Value string
}

type IotaAndOrEventFilter struct {
	Filter1 *IotaEventFilter
	Filter2 *IotaEventFilter
}

type TransactionFilter struct {
	ChangedObject *iotago.ObjectID
}

type IotaTransactionBlockEffects struct {
	V1 *IotaTransactionBlockEffectsV1
}

type IotaTransactionBlockEffectsV1 struct {
	Mutated []struct {
		Reference iotago.ObjectRef
	}
}

// IotaCoinMetadata holds metadata about a coin type (name, symbol, decimals, etc.).
type IotaCoinMetadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Description string
	IconURL     string
	ID          *iotago.ObjectID
}

// Supply holds the total supply of a coin type.
type Supply struct {
	Value *BigInt
}

// Re-export coin picking method constants.
const (
	PickMethodSmaller = graphqltypes.PickMethodSmaller
	PickMethodBigger  = graphqltypes.PickMethodBigger
	PickMethodByOrder = graphqltypes.PickMethodByOrder
)

// Re-export constants.
var (
	IotaCoinType = graphqltypes.IotaCoinType
)

// Re-export functions.
var (
	NewBigInt               = graphqltypes.NewBigInt
	NewBigIntInt64          = graphqltypes.NewBigIntInt64
	PickupCoins             = graphqltypes.PickupCoins
	PickupCoinsWithFilter   = graphqltypes.PickupCoinsWithFilter
	PickupCoinWithFilter    = graphqltypes.PickupCoinWithFilter
	PickupCoinsWithCointype = graphqltypes.PickupCoinsWithCointype
	MustCoinTypeFromString  = graphqltypes.MustCoinTypeFromString
	CoinTypeFromString      = graphqltypes.CoinTypeFromString
)
