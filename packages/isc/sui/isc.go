package sui

import (
	"github.com/howjmay/sui-go/sui_types"

	"github.com/iotaledger/wasp/packages/isc"
)

/*
  Structs are currently set up with the following assumptions:
  * Option<T> types are nullable types => (*T)
  * "ID" and "UID" are for now both typed as ObjectID, the actual typing maybe needs to be reconsidered. On our end it maybe not make a difference.
  * Type "Bag" is not available as a proper type in Sui-Go. It needs to be considered if we will need this, as
	=> Bag is a heterogeneous map, so it can hold key-value pairs of arbitrary types (map[any]any)
  * Type "Table" is a typed map: map[K]V
*/

// Related to: https://github.com/iotaledger/kinesis/tree/isc-models/dapps/isc/sources

type Allowance struct {
	CoinAmounts []uint64             `json:"coin_amounts"`
	CoinTypes   []string             `json:"coin_types"`
	NFTs        []sui_types.ObjectID `json:"nfts"`
}

type Assets struct {
	ID    sui_types.ObjectID `json:"id"`
	Coins Bag                `json:"coins"`
	NFTs  []NFT              `json:"nfts"`
}

type Anchor struct {
	ID sui_types.ObjectID `json:"id"`
}

type RequestData struct {
	ID        sui_types.ObjectID `json:"id"`
	Contract  isc.Hname          `json:"contract"`
	Function  isc.Hname          `json:"function"`
	Args      [][]uint8          `json:"args"`
	Allowance *Allowance         `json:"allowance"`
}

type Request struct {
	ID   sui_types.ObjectID `json:"id"`
	Data RequestData        `json:"data"`
}

// Related to: https://github.com/iotaledger/kinesis/blob/isc-models/crates/sui-framework/packages/stardust/sources/nft/irc27.move
type IRC27MetaData struct {
	Version           string                              `json:"version"`
	MediaType         string                              `json:"media_type"`
	URI               string                              `json:"uri"` // Actually of Type "Url" in SUI -> Create proper type?
	Name              string                              `json:"name"`
	CollectionName    *string                             `json:"collection_name"`
	Royalties         Table[sui_types.SuiAddress, uint32] `json:"royalties"`
	IssuerName        *string                             `json:"issuer_name"`
	Description       *string                             `json:"description"`
	Attributes        []string                            `json:"attributes"` // This is actually of Type VecSet which guarantees no duplicates. Not sure if we want to create a separate type for it. But we need to filter it to ensure no duplicates eventually.
	NonStandardFields Table[string, string]               `json:"non_standard_fields"`
}

// Related to: https://github.com/iotaledger/kinesis/blob/isc-models/crates/sui-framework/packages/stardust/sources/nft/nft.move

type NFT struct {
	ID                sui_types.ObjectID    `json:"id"`
	LegacySender      *sui_types.SuiAddress `json:"legacy_sender"`
	Metadata          *[]uint8              `json:"metadata"`
	Tag               *[]uint8              `json:"tag"`
	ImmutableIssuer   *sui_types.SuiAddress `json:"immutable_issuer"`
	ImmutableMetadata IRC27MetaData         `json:"immutable_metadata"`
}
