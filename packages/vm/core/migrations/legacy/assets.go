package legacy

import (
	"math/big"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

const (
	assetFlagHasBaseTokens   = 0x80
	assetFlagHasNativeTokens = 0x40
	assetFlagHasNFTs         = 0x20
)

/*
AssetsToBytes recreates the old Stardust assets encoding.
This should only ever be needed for the internal EVM Fake Transactions.
This works very well for BaseToken and NFTs, not for NativeToken.
This needs some validation regarding the compatibility between IDs.
*/
func AssetsToBytes(v isc.SchemaVersion, assets *isc.Assets) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return assets.Bytes()
	}

	w := rwutil.NewBytesWriter()

	if assets.IsEmpty() {
		return []byte{0x00}
	}

	hasBaseToken := assets.BaseTokens() > 0
	hasNFTs := !assets.Objects.IsEmpty()
	nts := assets.Coins.NativeTokens()
	hasNativeTokens := !nts.IsEmpty()

	var flags byte

	if hasBaseToken {
		flags |= assetFlagHasBaseTokens
	}
	if hasNativeTokens {
		flags |= assetFlagHasNativeTokens
	}
	if hasNFTs {
		flags |= assetFlagHasNFTs
	}

	w.WriteByte(flags)

	if (flags & assetFlagHasBaseTokens) != 0 {
		w.WriteAmount64(assets.BaseTokens().Uint64() / 1000)
	}

	if (flags & assetFlagHasNativeTokens) != 0 {
		w.WriteSize16(nts.Size())
		for t, v := range nts.Iterate() {
			w.WriteN(t.Bytes())
			n := v.BigInt().Div(v.BigInt(), new(big.Int).SetUint64(1000))
			w.WriteUint256(n)
		}
	}

	if (flags & assetFlagHasNFTs) != 0 {
		w.WriteSize16(assets.Objects.Size())
		for obj := range assets.Objects.Iterate() {
			w.WriteN(obj.ID[:])
		}
	}

	return w.Bytes()
}
