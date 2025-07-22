package iscmove

import (
	"bytes"
	"errors"
	"iter"
	"maps"
	"slices"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
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

// Hash returns the byte representation. Used in packages/chain/cons/bp/batch_proposal_set as key of a map
func (rwo *RefWithObject[any]) Hash() hashing.HashValue {
	return rwo.Digest.HashValue()
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
	if !a1.Equals(&a2.ObjectRef) {
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
	CoinBalances struct {
		items map[iotajsonrpc.CoinType]iotajsonrpc.CoinValue `bcs:"export"`
	}
	ObjectCollection struct {
		items map[iotago.ObjectID]iotago.ObjectType `bcs:"export"`
	}
)

func NewEmptyAssets() *Assets {
	return &Assets{
		Coins:   NewCoinBalances(),
		Objects: NewObjectCollection(),
	}
}

func NewAssets(baseTokens iotajsonrpc.CoinValue) *Assets {
	r := NewEmptyAssets()
	if baseTokens > 0 {
		r.SetCoin(iotajsonrpc.IotaCoinType, baseTokens)
	}
	return r
}

func NewCoinBalances() CoinBalances {
	return CoinBalances{
		items: make(map[iotajsonrpc.CoinType]iotajsonrpc.CoinValue),
	}
}

func (c CoinBalances) Set(coinType iotajsonrpc.CoinType, amount iotajsonrpc.CoinValue) {
	c.items[coinType] = amount
}

func (c CoinBalances) Get(coinType iotajsonrpc.CoinType) iotajsonrpc.CoinValue {
	return c.items[coinType]
}

// Iterate returns a deterministic iterator
func (c CoinBalances) Iterate() iter.Seq2[iotajsonrpc.CoinType, iotajsonrpc.CoinValue] {
	return func(yield func(iotajsonrpc.CoinType, iotajsonrpc.CoinValue) bool) {
		for _, k := range slices.Sorted(maps.Keys(c.items)) {
			if !yield(k, c.items[k]) {
				return
			}
		}
	}
}

func NewObjectCollection() ObjectCollection {
	return ObjectCollection{
		items: make(map[iotago.ObjectID]iotago.ObjectType),
	}
}

func (o ObjectCollection) Add(objectID iotago.ObjectID, t iotago.ObjectType) {
	o.items[objectID] = t
}

func (o ObjectCollection) Get(objectID iotago.ObjectID) (iotago.ObjectType, bool) {
	t, ok := o.items[objectID]
	return t, ok
}

func (o ObjectCollection) MustGet(objectID iotago.ObjectID) iotago.ObjectType {
	return o.items[objectID]
}

// Iterate returns a deterministic iterator
func (o ObjectCollection) Iterate() iter.Seq2[iotago.ObjectID, iotago.ObjectType] {
	return func(yield func(iotago.ObjectID, iotago.ObjectType) bool) {
		for _, k := range slices.SortedFunc(maps.Keys(o.items), func(a, b iotago.ObjectID) int {
			return bytes.Compare(a[:], b[:])
		}) {
			if !yield(k, o.items[k]) {
				return
			}
		}
	}
}

var ErrCoinNotFound = errors.New("coin not found")

func (a *Assets) FindCoin(coinType iotajsonrpc.CoinType) (iotajsonrpc.CoinValue, error) {
	for k, coin := range a.Coins.Iterate() {
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

func (a *Assets) SetCoin(coinType iotajsonrpc.CoinType, amount iotajsonrpc.CoinValue) *Assets {
	a.Coins.Set(coinType, amount)
	return a
}

func (a *Assets) AddObject(objectID iotago.ObjectID, t iotago.ObjectType) *Assets {
	a.Objects.Add(objectID, t)
	return a
}

func (a *Assets) BaseToken() iotajsonrpc.CoinValue {
	token, err := a.FindCoin(iotajsonrpc.IotaCoinType)
	if err != nil {
		if errors.Is(err, ErrCoinNotFound) {
			return 0
		}
		panic(err)
	}
	return token
}

type Request struct {
	ID        iotago.ObjectID
	Sender    *cryptolib.Address
	AssetsBag AssetsBagWithBalances
	Message   Message
	// AllowanceBCS is either empty or a BCS-encoded iscmove.Allowance
	AllowanceBCS []byte
	GasBudget    uint64
}

type RequestEvent struct {
	RequestID iotago.ObjectID
	Anchor    iotago.Address
}
