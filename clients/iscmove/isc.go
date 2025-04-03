package iscmove

import (
	"bytes"
	"errors"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	AnchorModuleName  = "anchor"
	AnchorObjectName  = "Anchor"
	ReceiptObjectName = "Receipt"

	AssetsBagModuleName = "assets_bag"
	AssetsBagObjectName = "AssetsBag"

	RequestModuleName      = "request"
	RequestObjectName      = "Request"
	MessageObjectName      = "Message"
	RequestEventObjectName = "RequestEvent"

	RequestEventAnchorFieldName = "/anchor"
)

/*
  Structs are currently set up with the following assumptions:
  * Option<T> types are nullable types => (*T)
    => Option<T> may require the `bcs:"optional"` tag.

  * "ID" and "UID" are for now both typed as ObjectID, the actual typing maybe needs to be reconsidered. On our end it maybe not make a difference.
  * Type "Bag" is not available as a proper type in Sui-Go. It needs to be considered if we will need this, as
	=> Bag is a heterogeneous map, so it can hold key-value pairs of arbitrary types (map[any]any)
  * Type "Table" is a typed map: map[K]V
*/

type RefWithObject[T any] struct {
	iotago.ObjectRef
	Object *T
	Owner  *iotago.Address
}

func init() {
	if hashing.HashSize != iotago.DigestSize {
		panic(errors.New("hashing.HashSize must be equal to iotago.DigestSize"))
	}
}

// Used in packages/chain/cons/bp/batch_proposal_set as key of a map
func (rwo *RefWithObject[any]) Hash() hashing.HashValue {
	return rwo.ObjectRef.Digest.HashValue()
}

type Referent[T any] struct {
	ID    iotago.ObjectID
	Value *T `bcs:"optional"`
}

// AssetsBag is the BCS equivalent for the move type AssetsBag
type AssetsBag struct {
	ID   iotago.ObjectID
	Size uint64
}

func (ab *AssetsBag) Equals(other *AssetsBag) bool {
	if (ab == nil) || (other == nil) {
		return (ab == nil) && (other == nil)
	}
	return ab.ID.Equals(other.ID) &&
		ab.Size == other.Size
}

type (
	AssetsBagWithBalances struct {
		AssetsBag
		Assets
	}
)

type Anchor struct {
	ID            iotago.ObjectID
	Assets        Referent[AssetsBag]
	StateMetadata []byte
	StateIndex    uint32
}

func (a1 Anchor) Equals(a2 *Anchor) bool {
	if !bytes.Equal(a1.ID[:], a2.ID[:]) {
		return false
	}
	if !bytes.Equal(a1.Assets.ID[:], a2.Assets.ID[:]) {
		return false
	}
	if !bytes.Equal(a1.Assets.Value.ID[:], a2.Assets.Value.ID[:]) {
		return false
	}
	if !a1.Assets.ID.Equals(a2.Assets.ID) {
		return false
	}
	if (a1.Assets.Value == nil) != (a2.Assets.Value == nil) {
		return false
	}
	if a1.Assets.Value != nil && !a1.Assets.Value.Equals(a2.Assets.Value) {
		return false
	}
	if !bytes.Equal(a1.StateMetadata, a2.StateMetadata) {
		return false
	}
	if a1.StateIndex != a2.StateIndex {
		return false
	}
	return true
}

type AnchorWithRef = RefWithObject[Anchor]

func AnchorWithRefEquals(a1 AnchorWithRef, a2 AnchorWithRef) bool {
	if !a1.ObjectRef.Equals(&a2.ObjectRef) {
		return false
	}
	if !a1.Object.Equals(a2.Object) {
		return false
	}
	return true
}

type Receipt struct {
	RequestID iotago.ObjectID
}

type Message struct {
	Contract uint32
	Function uint32
	Args     [][]byte
}

type Assets struct {
	Coins   CoinBalances
	Objects ObjectCollection
}

type (
	CoinBalances     map[iotajsonrpc.CoinType]iotajsonrpc.CoinValue
	ObjectCollection map[iotago.ObjectID]iotago.ObjectType
)

func NewEmptyAssets() *Assets {
	return &Assets{Coins: make(CoinBalances), Objects: make(ObjectCollection)}
}

func NewAssets(baseTokens uint64) *Assets {
	return NewEmptyAssets().AddCoin(iotajsonrpc.IotaCoinType, iotajsonrpc.CoinValue(baseTokens))
}

var ErrCoinNotFound = errors.New("coin not found")

func (a *Assets) FindCoin(coinType iotajsonrpc.CoinType) (iotajsonrpc.CoinValue, error) {
	for k, coin := range a.Coins {
		isSame, err := iotago.IsSameResource(k.String(), coinType.String())
		if err != nil {
			return 0, err
		}

		if isSame {
			return coin, nil
		}
	}

	return 0, ErrCoinNotFound
}

func (a *Assets) AddCoin(coinType iotajsonrpc.CoinType, amount iotajsonrpc.CoinValue) *Assets {
	a.Coins[coinType] = iotajsonrpc.CoinValue(amount)
	return a
}

func (a *Assets) AddObject(objectID iotago.ObjectID, t iotago.ObjectType) {
	a.Objects[objectID] = t
}

func (a *Assets) BaseToken() uint64 {
	token, err := a.FindCoin(iotajsonrpc.IotaCoinType)
	if err != nil {
		if errors.Is(err, ErrCoinNotFound) {
			return 0
		}

		panic(err)
	}

	return token.Uint64()
}

type Request struct {
	ID     iotago.ObjectID
	Sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
	AssetsBag AssetsBagWithBalances // Need to decide if we want to use this Referent wrapper as well. Could probably be of *AssetsBag with `bcs:"optional`
	Message   Message
	Allowance Assets
	GasBudget uint64
}

type RequestEvent struct {
	RequestID iotago.ObjectID
	Anchor    iotago.Address
}
