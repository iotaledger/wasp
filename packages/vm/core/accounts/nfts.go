package accounts

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/util"
)

func nftsMapKey(agentID isc.AgentID) string {
	return PrefixNFTs + string(agentID.Bytes())
}

func nftsByCollectionMapKey(agentID isc.AgentID, collectionKey kv.Key) string {
	return PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionKey)
}

func foundriesMapKey(agentID isc.AgentID) string {
	return PrefixFoundries + string(agentID.Bytes())
}

func accountToNFTsMapR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, nftsMapKey(agentID))
}

func accountToNFTsMap(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, nftsMapKey(agentID))
}

func NFTToOwnerMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyNFTOwner)
}

func NFTToOwnerMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyNFTOwner)
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
	return accountToNFTsMapR(state, agentID).HasAt(nftID[:])
}

func removeNFTOwner(state kv.KVStore, nftID iotago.NFTID, agentID isc.AgentID) bool {
	// remove the mapping of NFTID => owner
	nftMap := NFTToOwnerMap(state)
	if !nftMap.HasAt(nftID[:]) {
		return false
	}
	nftMap.DelAt(nftID[:])

	// add to the mapping of agentID => []NFTIDs
	nfts := accountToNFTsMap(state, agentID)
	if !nfts.HasAt(nftID[:]) {
		return false
	}
	nfts.DelAt(nftID[:])
	return true
}

func setNFTOwner(state kv.KVStore, nftID iotago.NFTID, agentID isc.AgentID) {
	// add to the mapping of NFTID => owner
	nftMap := NFTToOwnerMap(state)
	nftMap.SetAt(nftID[:], agentID.Bytes())

	// add to the mapping of agentID => []NFTIDs
	nfts := accountToNFTsMap(state, agentID)
	nfts.SetAt(nftID[:], codec.EncodeBool(true))
}

func GetNFTData(state kv.KVStoreReader, nftID iotago.NFTID) *isc.NFT {
	o, oID := GetNFTOutput(state, nftID)
	if o == nil {
		return nil
	}
	owner, err := isc.AgentIDFromBytes(NFTToOwnerMapR(state).GetAt(nftID[:]))
	if err != nil {
		panic("error parsing AgentID in NFTToOwnerMap")
	}
	return &isc.NFT{
		ID:       util.NFTIDFromNFTOutput(o, oID),
		Issuer:   o.ImmutableFeatureSet().IssuerFeature().Address,
		Metadata: o.ImmutableFeatureSet().MetadataFeature().Data,
		Owner:    owner,
	}
}

// CreditNFTToAccount credits an NFT to the on chain ledger
func CreditNFTToAccount(state kv.KVStore, agentID isc.AgentID, nftOutput *iotago.NFTOutput, chainID isc.ChainID) {
	if nftOutput.NFTID.Empty() {
		panic("empty NFTID")
	}

	creditNFTToAccount(state, agentID, nftOutput.NFTID, nftOutput.ImmutableFeatureSet().IssuerFeature().Address)
	touchAccount(state, agentID, chainID)

	// save the NFTOutput with a temporary outputIndex so the NFTData is readily available (it will be updated upon block closing)
	SaveNFTOutput(state, nftOutput, 0)
}

func creditNFTToAccount(state kv.KVStore, agentID isc.AgentID, nftID iotago.NFTID, issuer iotago.Address) {
	setNFTOwner(state, nftID, agentID)

	collectionKey := nftCollectionKey(issuer)
	nftsByCollection := nftsByCollectionMap(state, agentID, collectionKey)
	nftsByCollection.SetAt(nftID[:], codec.EncodeBool(true))
}

// DebitNFTFromAccount removes an NFT from an account.
// If the account does not own the nft, it panics.
func DebitNFTFromAccount(state kv.KVStore, agentID isc.AgentID, nftID iotago.NFTID, chainID isc.ChainID) {
	nft := GetNFTData(state, nftID)
	if nft == nil {
		panic(fmt.Errorf("cannot debit unknown NFT %s", nftID.String()))
	}
	if !debitNFTFromAccount(state, agentID, nft) {
		panic(fmt.Errorf("cannot debit NFT %s from %s: %w", nftID.String(), agentID, ErrNotEnoughFunds))
	}
	touchAccount(state, agentID, chainID)
}

// DebitNFTFromAccount removes an NFT from the internal map of an account
func debitNFTFromAccount(state kv.KVStore, agentID isc.AgentID, nft *isc.NFT) bool {
	if !removeNFTOwner(state, nft.ID, agentID) {
		return false
	}

	collectionKey := nftCollectionKey(nft.Issuer)
	nftsByCollection := nftsByCollectionMap(state, agentID, collectionKey)
	if !nftsByCollection.HasAt(nft.ID[:]) {
		panic("inconsistency: NFT not found in collection")
	}
	nftsByCollection.DelAt(nft.ID[:])

	return true
}

func collectNFTIDs(m *collections.ImmutableMap) []iotago.NFTID {
	var ret []iotago.NFTID
	m.Iterate(func(idBytes []byte, val []byte) bool {
		id := iotago.NFTID{}
		copy(id[:], idBytes)
		ret = append(ret, id)
		return true
	})
	return ret
}

func getAccountNFTs(state kv.KVStoreReader, agentID isc.AgentID) []iotago.NFTID {
	return collectNFTIDs(accountToNFTsMapR(state, agentID))
}

func getAccountNFTsInCollection(state kv.KVStoreReader, agentID isc.AgentID, collectionID iotago.NFTID) []iotago.NFTID {
	return collectNFTIDs(nftsByCollectionMapR(state, agentID, kv.Key(collectionID[:])))
}

func getL2TotalNFTs(state kv.KVStoreReader) []iotago.NFTID {
	return collectNFTIDs(NFTToOwnerMapR(state))
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
	allAccountsMapR(state).IterateKeys(func(key []byte) bool {
		agentID, err := isc.AgentIDFromBytes(key)
		if err != nil {
			panic(fmt.Errorf("calcL2TotalNFTs: %w", err))
		}
		ret = append(ret, getAccountNFTs(state, agentID)...)
		return true
	})
	return ret
}
