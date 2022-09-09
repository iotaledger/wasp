package accounts

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

// CreditNFTToAccount credits an NFT to the on chain ledger
func CreditNFTToAccount(state kv.KVStore, agentID isc.AgentID, nft *isc.NFT) {
	if nft == nil {
		return
	}
	account := getAccount(state, agentID)

	checkLedger(state, "CreditNFTToAccount IN")
	defer checkLedger(state, "CreditNFTToAccount OUT")

	saveNFTData(state, nft)
	creditNFTToAccount(state, account, nft.ID, agentID)
	touchAccount(state, account)
}

func saveNFTData(state kv.KVStore, nft *isc.NFT) {
	nftMap := getNFTState(state)
	if nftMap.MustHasAt(nft.ID[:]) {
		panic("saveNFTData: inconsistency - NFT data already exists")
	}
	nftMap.MustSetAt(nft.ID[:], nft.Bytes(false))
}

func deleteNFTData(state kv.KVStore, id iotago.NFTID) {
	nftMap := getNFTState(state)
	if !nftMap.MustHasAt(id[:]) {
		panic("deleteNFTData: inconsistency - NFT data doesn't exists")
	}
	nftMap.MustDelAt(id[:])
}

func GetNFTData(state kv.KVStoreReader, id iotago.NFTID) isc.NFT {
	nftMap := getNFTStateR(state)
	nftBytes := nftMap.MustGetAt(id[:])
	if len(nftBytes) == 0 {
		panic(ErrNFTIDNotFound.Create(id))
	}
	nft, err := isc.NFTFromBytes(nftBytes, false)
	nft.ID = id
	if err != nil {
		panic(fmt.Sprintf("getNFTData: error when parsing NFTdata: %v", err))
	}
	if nft == nil {
		panic(ErrNFTIDNotFound.Create(id))
	}
	return *nft
}

func creditNFTToAccount(state kv.KVStore, account *collections.Map, id iotago.NFTID, agentID isc.AgentID) {
	account.MustSetAt(id[:], codec.EncodeBool(true))
	nftMap := getNFTState(state)
	nftBytes := nftMap.MustGetAt(id[:])
	nft, err := isc.NFTFromBytes(nftBytes, false)
	if err != nil {
		panic(fmt.Sprintf("creditNFTToAccount: error when parsing NFTdata: %v", err))
	}
	nft.Owner = agentID
	nftMap.MustSetAt(id[:], nft.Bytes(false))
}

// DebitNFTFromAccount removes an NFT from an account. if that account doesn't own the nft, it panics
// this will also delete the NFT data, as the NFT will be leaving the chain
func DebitNFTFromAccount(state kv.KVStore, agentID isc.AgentID, id iotago.NFTID) {
	if id.Empty() {
		return
	}
	account := getAccount(state, agentID)

	checkLedger(state, "DebitNFTFromAccount IN")
	defer checkLedger(state, "DebitNFTFromAccount OUT")

	if !debitNFTFromAccount(account, id) {
		panic(xerrors.Errorf(" debit NFT from %s: %v\nassets: %s", agentID, ErrNotEnoughFunds, id.String()))
	}

	deleteNFTData(state, id)
	touchAccount(state, account)
}

// DebitNFTFromAccount removes an NFT from the internal map of an account
func debitNFTFromAccount(account *collections.Map, id iotago.NFTID) bool {
	bytes, err := account.GetAt(id[:])
	if err != nil || len(bytes) == 0 {
		return false
	}
	err = account.DelAt(id[:])
	return err == nil
}
