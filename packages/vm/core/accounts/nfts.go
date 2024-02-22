package accounts

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/util"
)

// observation: this uses the entire agentID as key, unlike acccounts.accountKey, which skips the chainID if it is the current chain. This means some bytes are wasted when saving NFTs

func nftsMapKey(agentID isc.AgentID) string {
	return PrefixNFTs + string(agentID.Bytes())
}

func nftsByCollectionMapKey(agentID isc.AgentID, collectionKey kv.Key) string {
	return PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionKey)
}

func foundriesMapKey(agentID isc.AgentID) string {
	return PrefixFoundries + string(agentID.Bytes())
}

func (s *StateReader) accountToNFTsMap(agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, nftsMapKey(agentID))
}

func (s *StateWriter) accountToNFTsMap(agentID isc.AgentID) *collections.Map {
	return collections.NewMap(s.state, nftsMapKey(agentID))
}

func (s *StateWriter) nftToOwnerMap() *collections.Map {
	return collections.NewMap(s.state, keyNFTOwner)
}

func (s *StateReader) nftToOwnerMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyNFTOwner)
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

func (s *StateReader) nftsByCollectionMap(agentID isc.AgentID, collectionKey kv.Key) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, nftsByCollectionMapKey(agentID, collectionKey))
}

func (s *StateWriter) nftsByCollectionMap(agentID isc.AgentID, collectionKey kv.Key) *collections.Map {
	return collections.NewMap(s.state, nftsByCollectionMapKey(agentID, collectionKey))
}

func (s *StateReader) hasNFT(agentID isc.AgentID, nftID iotago.NFTID) bool {
	return s.accountToNFTsMap(agentID).HasAt(nftID[:])
}

func (s *StateWriter) removeNFTOwner(nftID iotago.NFTID, agentID isc.AgentID) bool {
	// remove the mapping of NFTID => owner
	nftMap := s.nftToOwnerMap()
	if !nftMap.HasAt(nftID[:]) {
		return false
	}
	nftMap.DelAt(nftID[:])

	// add to the mapping of agentID => []NFTIDs
	nfts := s.accountToNFTsMap(agentID)
	if !nfts.HasAt(nftID[:]) {
		return false
	}
	nfts.DelAt(nftID[:])
	return true
}

func (s *StateWriter) setNFTOwner(nftID iotago.NFTID, agentID isc.AgentID) {
	// add to the mapping of NFTID => owner
	nftMap := s.nftToOwnerMap()
	nftMap.SetAt(nftID[:], agentID.Bytes())

	// add to the mapping of agentID => []NFTIDs
	nfts := s.accountToNFTsMap(agentID)
	nfts.SetAt(nftID[:], codec.Bool.Encode(true))
}

func (s *StateReader) GetNFTData(nftID iotago.NFTID) *isc.NFT {
	o, oID := s.GetNFTOutput(nftID)
	if o == nil {
		return nil
	}
	owner, err := isc.AgentIDFromBytes(s.nftToOwnerMap().GetAt(nftID[:]))
	if err != nil {
		panic("error parsing AgentID in NFTToOwnerMap")
	}
	nft := &isc.NFT{
		ID:    util.NFTIDFromNFTOutput(o, oID),
		Owner: owner,
	}
	if issuer := o.ImmutableFeatureSet().Issuer(); issuer != nil {
		nft.Issuer = issuer.Address
	}
	if md := o.ImmutableFeatureSet().Metadata(); md != nil {
		nft.Metadata = md.Entries
	}
	return nft
}

// CreditNFTToAccount credits an NFT to the on chain ledger
func (s *StateWriter) CreditNFTToAccount(agentID isc.AgentID, nftOutput *iotago.NFTOutput) {
	if nftOutput.NFTID.Empty() {
		panic("empty NFTID")
	}

	var issuer iotago.Address
	issuerFeature := nftOutput.ImmutableFeatureSet().Issuer()
	if issuerFeature != nil {
		issuer = issuerFeature.Address
	}
	s.creditNFTToAccount(agentID, nftOutput.NFTID, issuer)
	s.touchAccount(agentID)

	// save the NFTOutput with a temporary outputIndex so the NFTData is readily available (it will be updated upon block closing)
	s.SaveNFTOutput(nftOutput, 0)
}

func (s *StateWriter) creditNFTToAccount(agentID isc.AgentID, nftID iotago.NFTID, issuer iotago.Address) {
	s.setNFTOwner(nftID, agentID)

	collectionKey := nftCollectionKey(issuer)
	nftsByCollection := s.nftsByCollectionMap(agentID, collectionKey)
	nftsByCollection.SetAt(nftID[:], codec.Bool.Encode(true))
}

// DebitNFTFromAccount removes an NFT from an account.
// If the account does not own the nft, it panics.
func (s *StateWriter) DebitNFTFromAccount(agentID isc.AgentID, nftID iotago.NFTID) {
	nft := s.GetNFTData(nftID)
	if nft == nil {
		panic(fmt.Errorf("cannot debit unknown NFT %s", nftID.String()))
	}
	if !s.debitNFTFromAccount(agentID, nft) {
		panic(fmt.Errorf("cannot debit NFT %s from %s: %w", nftID.String(), agentID, ErrNotEnoughFunds))
	}
	s.touchAccount(agentID)
}

// DebitNFTFromAccount removes an NFT from the internal map of an account
func (s *StateWriter) debitNFTFromAccount(agentID isc.AgentID, nft *isc.NFT) bool {
	if !s.removeNFTOwner(nft.ID, agentID) {
		return false
	}

	collectionKey := nftCollectionKey(nft.Issuer)
	nftsByCollection := s.nftsByCollectionMap(agentID, collectionKey)
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

func (s *StateReader) getAccountNFTs(agentID isc.AgentID) []iotago.NFTID {
	return collectNFTIDs(s.accountToNFTsMap(agentID))
}

func (s *StateReader) getAccountNFTsInCollection(agentID isc.AgentID, collectionID iotago.NFTID) []iotago.NFTID {
	return collectNFTIDs(s.nftsByCollectionMap(agentID, kv.Key(collectionID[:])))
}

func (s *StateReader) getL2TotalNFTs() []iotago.NFTID {
	return collectNFTIDs(s.nftToOwnerMap())
}

// GetAccountNFTs returns all NFTs belonging to the agentID on the state
func (s *StateReader) GetAccountNFTs(agentID isc.AgentID) []iotago.NFTID {
	return s.getAccountNFTs(agentID)
}

func (s *StateReader) GetTotalL2NFTs() []iotago.NFTID {
	return s.getL2TotalNFTs()
}
