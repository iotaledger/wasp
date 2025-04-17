package legacy

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

const (
	assetFlagHasBaseTokens   = 0x80
	assetFlagHasNativeTokens = 0x40
	assetFlagHasNFTs         = 0x20
)

/*
Recreating the old Stardust assets encoding. This should only ever be needed for the internal EVM Fake Transactions.
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
	hasNFTs := len(assets.Objects) > 0
	hasNativeTokens := len(assets.Coins.NativeTokens()) > 0

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
		nts := assets.Coins.NativeTokens()
		w.WriteSize16(len(nts))
		for t, v := range nts.Iterate() {
			w.WriteN(t.Bytes()[:])
			n := v.BigInt().Div(v.BigInt(), new(big.Int).SetUint64(1000))
			w.WriteUint256(n)
		}
	}

	if (flags & assetFlagHasNFTs) != 0 {
		w.WriteSize16(len(assets.Objects))
		for obj := range assets.Objects.Iterate() {
			w.WriteN(obj.ID[:])
		}
	}

	return w.Bytes()
}
