package iscmove

import (
	"bytes"

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

// Used in packages/chain/cons/bp/batch_proposal_set as key of a map
// TODO: maybe use a.Ref.Digest() instead? Maybe have Key() for RefWithObject type?
func (rwo *RefWithObject[any]) Hash() hashing.HashValue {
	res, _ := hashing.HashValueFromBytes(rwo.ObjectRef.Bytes())
	return res
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

type AssetsBagBalances map[iotajsonrpc.CoinType]*iotajsonrpc.Balance

type AssetsBagWithBalances struct {
	AssetsBag
	Balances AssetsBagBalances `bcs:"-"`
}

type Anchor struct {
	ID            iotago.ObjectID
	Assets        AssetsBag
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
	if !bytes.Equal(a1.Assets.ID[:], a2.Assets.ID[:]) {
		return false
	}
	if !bytes.Equal(a1.Assets.ID[:], a2.Assets.ID[:]) {
		return false
	}
	if a1.Assets.Size != a2.Assets.Size {
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
	Coins CoinBalances
}

type CoinAllowance struct {
	CoinType iotajsonrpc.CoinType
	Balance  iotajsonrpc.CoinValue
}

type CoinBalances map[iotajsonrpc.CoinType]iotajsonrpc.CoinValue

func NewEmptyAssets() *Assets {
	return &Assets{Coins: make(CoinBalances)}
}

func NewAssets(baseTokens iotajsonrpc.CoinValue) *Assets {
	return NewEmptyAssets().AddCoin(iotajsonrpc.IotaCoinType, baseTokens)
}

func (a *Assets) AddCoin(coinType iotajsonrpc.CoinType, amount iotajsonrpc.CoinValue) *Assets {
	a.Coins[coinType] = iotajsonrpc.CoinValue(amount)
	return a
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

// Related to: https://github.com/iotaledger/kinesis/blob/isc-iotajsonrpc/crates/sui-framework/packages/stardust/sources/nft/irc27.move
type IRC27MetaData struct {
	Version           string
	MediaType         string
	URI               string
	Name              string
	CollectionName    *string `bcs:"optional"`
	Royalties         Table[*cryptolib.Address, uint32]
	IssuerName        *string  `bcs:"optional"`
	Description       *string  `bcs:"optional"`
	Attributes        []string // This is actually of Type VecSet which guarantees no duplicates. Not sure if we want to create a separate type for it. But we need to filter it to ensure no duplicates eventually.
	NonStandardFields Table[string, string]
}

type NFT struct {
	ID                iotago.ObjectID
	LegacySender      *cryptolib.Address `bcs:"optional"`
	Metadata          *[]uint8           `bcs:"optional"`
	Tag               *[]uint8           `bcs:"optional"`
	ImmutableIssuer   *cryptolib.Address `bcs:"optional"`
	ImmutableMetadata IRC27MetaData
}
