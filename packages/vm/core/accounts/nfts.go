package accounts

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func nftsMapKey(agentID isc.AgentID) string {
	return prefixNFTs + string(agentID.Bytes())
}

func nftsByCollectionMapKey(agentID isc.AgentID, collectionKey kv.Key) string {
	return prefixNFTsByCollection + string(agentID.Bytes()) + string(collectionKey)
}

func foundriesMapKey(agentID isc.AgentID) string {
	return prefixFoundries + string(agentID.Bytes())
}

func nftsMapR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, nftsMapKey(agentID))
}

func nftsMap(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, nftsMapKey(agentID))
}

func nftDataMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyNFTData)
}

func nftDataMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyNFTData)
}

func nftCollectionKey(issuer iotago.Address) kv.Key {
	if issuer == nil {
		return noCollection
	}
	nftAddr, ok := issuer.(*iotago.NFTAddress)
	if !ok {
		return noCollection
	}
	id := nftAddr.NFTID()
	return kv.Key(id[:])
}

func nftsByCollectionMapR(state kv.KVStoreReader, agentID isc.AgentID, collectionKey kv.Key) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, nftsByCollectionMapKey(agentID, collectionKey))
}

func nftsByCollectionMap(state kv.KVStore, agentID isc.AgentID, collectionKey kv.Key) *collections.Map {
	return collections.NewMap(state, nftsByCollectionMapKey(agentID, collectionKey))
}

func hasNFT(state kv.KVStoreReader, agentID isc.AgentID, nftID iotago.NFTID) bool {
	return nftsMapR(state, agentID).MustHasAt(nftID[:])
}

func saveNFTData(state kv.KVStore, nft *isc.NFT) {
	nftDataMap(state).MustSetAt(nft.ID[:], nft.Bytes(false))
}

func deleteNFTData(state kv.KVStore, id iotago.NFTID) {
	allNFTs := nftDataMap(state)
	if !allNFTs.MustHasAt(id[:]) {
		panic("deleteNFTData: inconsistency - NFT data doesn't exists")
	}
	allNFTs.MustDelAt(id[:])
}

func GetNFTData(state kv.KVStoreReader, id iotago.NFTID) (*isc.NFT, error) {
	allNFTs := nftDataMapR(state)
	b := allNFTs.MustGetAt(id[:])
	if len(b) == 0 {
		return nil, ErrNFTIDNotFound.Create(id)
	}
	nft, err := isc.NFTFromBytes(b, false)
	if err != nil {
		panic(fmt.Sprintf("getNFTData: error when parsing NFTdata: %v", err))
	}
	nft.ID = id
	return nft, nil
}

func MustGetNFTData(state kv.KVStoreReader, id iotago.NFTID) *isc.NFT {
	nft, err := GetNFTData(state, id)
	if err != nil {
		panic(err)
	}
	return nft
}

// CreditNFTToAccount credits an NFT to the on chain ledger
func CreditNFTToAccount(state kv.KVStore, agentID isc.AgentID, nft *isc.NFT) {
	if nft == nil {
		return
	}
	if nft.ID.Empty() {
		panic("empty NFTID")
	}

	creditNFTToAccount(state, agentID, nft)
	touchAccount(state, agentID)
}

func creditNFTToAccount(state kv.KVStore, agentID isc.AgentID, nft *isc.NFT) {
	nft.Owner = agentID
	saveNFTData(state, nft)

	nfts := nftsMap(state, agentID)
	nfts.MustSetAt(nft.ID[:], codec.EncodeBool(true))

	collectionKey := nftCollectionKey(nft.Issuer)
	nftsByCollection := nftsByCollectionMap(state, agentID, collectionKey)
	nftsByCollection.MustSetAt(nft.ID[:], codec.EncodeBool(true))
}

// DebitNFTFromAccount removes an NFT from an account.
// If the account does not own the nft, it panics.
func DebitNFTFromAccount(state kv.KVStore, agentID isc.AgentID, id iotago.NFTID) {
	nft, err := GetNFTData(state, id)
	if err != nil {
		panic(err)
	}
	if !debitNFTFromAccount(state, agentID, nft) {
		panic(fmt.Errorf(" debit NFT from %s: %v\nassets: %s", agentID, ErrNotEnoughFunds, id.String()))
	}
	deleteNFTData(state, id)
	touchAccount(state, agentID)
}

// DebitNFTFromAccount removes an NFT from the internal map of an account
func debitNFTFromAccount(state kv.KVStore, agentID isc.AgentID, nft *isc.NFT) bool {
	nfts := nftsMap(state, agentID)
	if !nfts.MustHasAt(nft.ID[:]) {
		return false
	}
	nfts.MustDelAt(nft.ID[:])

	collectionKey := nftCollectionKey(nft.Issuer)
	nftsByCollection := nftsByCollectionMap(state, agentID, collectionKey)
	if !nftsByCollection.MustHasAt(nft.ID[:]) {
		panic("inconsistency: NFT not found in collection")
	}
	nftsByCollection.MustDelAt(nft.ID[:])

	return true
}

func collectNFTIDs(m *collections.ImmutableMap) []iotago.NFTID {
	var ret []iotago.NFTID
	m.MustIterate(func(idBytes []byte, val []byte) bool {
		id := iotago.NFTID{}
		copy(id[:], idBytes)
		ret = append(ret, id)
		return true
	})
	return ret
}

func getAccountNFTs(state kv.KVStoreReader, agentID isc.AgentID) []iotago.NFTID {
	return collectNFTIDs(nftsMapR(state, agentID))
}

func getAccountNFTsInCollection(state kv.KVStoreReader, agentID isc.AgentID, collectionID iotago.NFTID) []iotago.NFTID {
	return collectNFTIDs(nftsByCollectionMapR(state, agentID, kv.Key(collectionID[:])))
}

func getL2TotalNFTs(state kv.KVStoreReader) []iotago.NFTID {
	return collectNFTIDs(nftDataMapR(state))
}

// GetAccountNFTs returns all NFTs belonging to the agentID on the state
func GetAccountNFTs(state kv.KVStoreReader, agentID isc.AgentID) []iotago.NFTID {
	return getAccountNFTs(state, agentID)
}

func GetTotalL2NFTs(state kv.KVStoreReader) []iotago.NFTID {
	return getL2TotalNFTs(state)
}

func calcL2TotalNFTs(state kv.KVStoreReader) []iotago.NFTID {
	var ret []iotago.NFTID
	allAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := isc.AgentIDFromBytes(key)
		if err != nil {
			panic(fmt.Errorf("calcL2TotalNFTs: %w", err))
		}
		ret = append(ret, getAccountNFTs(state, agentID)...)
		return true
	})
	return ret
}
