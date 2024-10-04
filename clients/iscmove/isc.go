package iscmove

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

const (
	AnchorModuleName  = "anchor"
	AnchorObjectName  = "Anchor"
	ReceiptObjectName = "Receipt"

	AssetsBagModuleName = "assets_bag"
	AssetsBagObjectName = "AssetsBag"
	AssetObjectName     = "Asset"

	RequestModuleName      = "request"
	RequestObjectName      = "Request"
	MessageObjectName      = "Message"
	RequestEventObjectName = "RequestEvent"
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

// Related to: https://github.com/iotaledger/kinesis/tree/isc-suijsonrpc/dapps/isc/sources
// Might change completely: https://github.com/iotaledger/iota/pull/370#discussion_r1617682560
type AllowanceWithVales struct {
	CoinAmounts []uint64
	CoinTypes   []suijsonrpc.CoinType
	NFTs        []sui.ObjectID
}

type Allowance struct {
	ID sui.ObjectID
}

type Referent[T any] struct {
	ID    sui.ObjectID
	Value *T `bcs:"optional"`
}

type RefWithObject[T any] struct {
	sui.ObjectRef
	Object *T
}

// Used in packages/chain/cons/bp/batch_proposal_set as key of a map
// TODO: maybe use a.Ref.Digest() instead? Maybe have Key() for RefWithObject type?
func (rwo *RefWithObject[any]) Hash() hashing.HashValue {
	res, _ := hashing.HashValueFromBytes(rwo.ObjectRef.Bytes())
	return res
}

// AssetsBag is the BCS equivalent for the move type AssetsBag
type AssetsBag struct {
	ID   sui.ObjectID
	Size uint64
}

type AssetsBagBalances map[suijsonrpc.CoinType]*suijsonrpc.Balance

type AssetsBagWithBalances struct {
	AssetsBag
	Balances AssetsBagBalances `bcs:"-"`
}

// Anchor is the BCS equivalent for the move type Anchor
type Anchor struct {
	ID            sui.ObjectID
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
	if !bytes.Equal(a1.Assets.ID[:], a2.Assets.ID[:]) {
		return false
	}
	if !bytes.Equal(a1.Assets.Value.ID[:], a2.Assets.Value.ID[:]) {
		return false
	}
	if a1.Assets.Value.Size != a2.Assets.Value.Size {
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
	RequestID sui.ObjectID
}

type Message struct {
	Contract uint32
	Function uint32
	Args     [][]byte
}

type CoinAllowance struct {
	CoinType string
	Balance  uint64
}

type Request struct {
	ID     sui.ObjectID
	Sender *cryptolib.Address
	// XXX balances are empty if we don't fetch the dynamic fields
	AssetsBag Referent[AssetsBagWithBalances] // Need to decide if we want to use this Referent wrapper as well. Could probably be of *AssetsBag with `bcs:"optional`
	Message   Message
	Allowance []CoinAllowance
	GasBudget uint64
}

// Related to: https://github.com/iotaledger/kinesis/blob/isc-suijsonrpc/crates/sui-framework/packages/stardust/sources/nft/irc27.move
type IRC27MetaData struct {
	Version           string
	MediaType         string
	URI               string // Actually of type "Url" in SUI -> Create proper type?
	Name              string
	CollectionName    *string `bcs:"optional"`
	Royalties         Table[*cryptolib.Address, uint32]
	IssuerName        *string  `bcs:"optional"`
	Description       *string  `bcs:"optional"`
	Attributes        []string // This is actually of Type VecSet which guarantees no duplicates. Not sure if we want to create a separate type for it. But we need to filter it to ensure no duplicates eventually.
	NonStandardFields Table[string, string]
}

type NFT struct {
	ID                sui.ObjectID
	LegacySender      *cryptolib.Address `bcs:"optional"`
	Metadata          *[]uint8           `bcs:"optional"`
	Tag               *[]uint8           `bcs:"optional"`
	ImmutableIssuer   *cryptolib.Address `bcs:"optional"`
	ImmutableMetadata IRC27MetaData
}
