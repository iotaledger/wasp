package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func nftOutputMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyNFTOutputRecords)
}

func nftOutputMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyNFTOutputRecords)
}

// SaveNFTOutput map tokenID -> foundryRec
func SaveNFTOutput(state kv.KVStore, out *iotago.NFTOutput, blockIndex uint32, outputIndex uint16) {
	tokenRec := NFTOutputRec{
		Output:      out,
		BlockIndex:  blockIndex,
		OutputIndex: outputIndex,
	}
	nftOutputMap(state).MustSetAt(out.NFTID[:], tokenRec.Bytes())
}

func DeleteNFTOutput(state kv.KVStore, id iotago.NFTID) {
	nftOutputMap(state).MustDelAt(id[:])
}

func GetNFTOutput(state kv.KVStoreReader, id iotago.NFTID, chainID isc.ChainID) (*iotago.NFTOutput, uint32, uint16) {
	data := nftOutputMapR(state).MustGetAt(id[:])
	if data == nil {
		return nil, 0, 0
	}
	tokenRec := mustNFTOutputRecFromBytes(data)
	return tokenRec.Output, tokenRec.BlockIndex, tokenRec.OutputIndex
}
