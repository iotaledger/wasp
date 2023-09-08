package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func newNFTsArray(state kv.KVStore) *collections.Array {
	return collections.NewArray(state, keyNewNFTs)
}

func NFTOutputMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyNFTOutputRecords)
}

func nftOutputMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyNFTOutputRecords)
}

func SaveNFTOutput(state kv.KVStore, out *iotago.NFTOutput, outputIndex uint16) {
	tokenRec := NFTOutputRec{
		// TransactionID is unknown yet, will be filled next block
		OutputID: iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, outputIndex),
		Output:   out,
	}
	NFTOutputMap(state).SetAt(out.NFTID[:], tokenRec.Bytes())
	newNFTsArray(state).Push(out.NFTID[:])
}

func updateNFTOutputIDs(state kv.KVStore, anchorTxID iotago.TransactionID) {
	newNFTs := newNFTsArray(state)
	allNFTs := NFTOutputMap(state)
	n := newNFTs.Len()
	for i := uint32(0); i < n; i++ {
		nftID := newNFTs.GetAt(i)
		rec := mustNFTOutputRecFromBytes(allNFTs.GetAt(nftID))
		rec.OutputID = iotago.OutputIDFromTransactionIDAndIndex(anchorTxID, rec.OutputID.Index())
		allNFTs.SetAt(nftID, rec.Bytes())
	}
	newNFTs.Erase()
}

func DeleteNFTOutput(state kv.KVStore, nftID iotago.NFTID) {
	NFTOutputMap(state).DelAt(nftID[:])
}

func GetNFTOutput(state kv.KVStoreReader, nftID iotago.NFTID) (*iotago.NFTOutput, iotago.OutputID) {
	data := nftOutputMapR(state).GetAt(nftID[:])
	if data == nil {
		return nil, iotago.OutputID{}
	}
	tokenRec := mustNFTOutputRecFromBytes(data)
	return tokenRec.Output, tokenRec.OutputID
}
