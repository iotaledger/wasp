package m002

import (
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

// for testnet -- delete when deploying ShimmerEVM
var DeprecateNFTData = migrations.Migration{
	Contract: accounts.Contract,
	Apply: func(state kv.KVStore, log *logger.Logger) error {
		oldNFTDataMap := collections.NewMap(state, "ND")
		nftToOwnerMap := collections.NewMap(state, "NW")
		oldNFTDataMap.Iterate(func(nftIDBytes, nftDataBytes []byte) bool {
			rr := rwutil.NewBytesReader(nftDataBytes)
			// note we stored the NFT data without the leading id bytes
			rr.PushBack().WriteN(nftIDBytes)
			nft, err := isc.NFTFromReader(rr)
			if err != nil {
				panic(err)
			}
			if nft.Owner == nil {
				log.Errorf("DeprecateNFTData migration | nil owner at NFTID: %s", iotago.EncodeHex(nftIDBytes))
			}
			nftToOwnerMap.SetAt(nftIDBytes, nft.Owner.Bytes())
			return true
		})
		return nil
	},
}
