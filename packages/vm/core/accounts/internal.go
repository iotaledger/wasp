package accounts

import (
	"errors"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var (
	ErrNotEnoughFunds                       = coreerrors.Register("not enough funds").Create()
	ErrNotEnoughBaseTokensForStorageDeposit = coreerrors.Register("not enough base tokens for storage deposit").Create()
	ErrNotEnoughAllowance                   = coreerrors.Register("not enough allowance").Create()
	ErrBadAmount                            = coreerrors.Register("bad native asset amount").Create()
	ErrRepeatingFoundrySerialNumber         = coreerrors.Register("repeating serial number of the foundry").Create()
	ErrFoundryNotFound                      = coreerrors.Register("foundry not found").Create()
	ErrOverflow                             = coreerrors.Register("overflow in token arithmetics").Create()
	ErrTooManyNFTsInAllowance               = coreerrors.Register("expected at most 1 NFT in allowance").Create()
	ErrNFTIDNotFound                        = coreerrors.Register("NFTID not found").Create()
)

const (
	// keyAllAccounts stores a map of <agentID> => true
	// where sum = baseTokens + native tokens + nfts
	keyAllAccounts = "a"

	// prefixBaseTokens | <accountID> stores the amount of base tokens (big.Int)
	prefixBaseTokens = "b"
	// prefixBaseTokens | <accountID> stores a map of <nativeTokenID> => big.Int
	PrefixNativeTokens = "t"

	// l2TotalsAccount is the special <accountID> storing the total fungible tokens
	// controlled by the chain
	l2TotalsAccount = "*"

	// PrefixNFTs | <agentID> stores a map of <NFTID> => true
	PrefixNFTs = "n"
	// PrefixNFTsByCollection | <agentID> | <collectionID> stores a map of <nftID> => true
	PrefixNFTsByCollection = "c"
	// prefixNewlyMintedNFTs stores a map of <position in minted list> => <newly minted NFT> to be updated when the outputID is known
	prefixNewlyMintedNFTs = "N"
	// prefixMintIDMap stores a map of <internal NFTID> => <NFTID> it is updated when the NFTID of newly minted nfts is known
	prefixMintIDMap = "M"
	// PrefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	PrefixFoundries = "f"

	// noCollection is the special <collectionID> used for storing NFTs that do not belong in a collection
	noCollection = "-"

	// keyNonce stores a map of <agentID> => nonce (uint64)
	keyNonce = "m"

	// keyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	keyNativeTokenOutputMap = "TO"
	// keyFoundryOutputRecords stores a map of <foundrySN> => foundryOutputRec
	keyFoundryOutputRecords = "FO"
	// keyNFTOutputRecords stores a map of <NFTID> => NFTOutputRec
	keyNFTOutputRecords = "NO"
	// keyNFTOwner stores a map of <NFTID> => isc.AgentID
	keyNFTOwner = "NW"

	// keyNewNativeTokens stores an array of <nativeTokenID>, containing the newly created native tokens that need filling out the OutputID
	keyNewNativeTokens = "TN"
	// keyNewFoundries stores an array of <foundrySN>, containing the newly created foundries that need filling out the OutputID
	keyNewFoundries = "FN"
	// keyNewNFTs stores an array of <NFTID>, containing the newly created NFTs that need filling out the OutputID
	keyNewNFTs = "NN"
)

func accountKey(agentID isc.AgentID, chainID isc.ChainID) kv.Key {
	if agentID.BelongsToChain(chainID) {
		// save bytes by skipping the chainID bytes on agentIDs for this chain
		return kv.Key(agentID.BytesWithoutChainID())
	}
	return kv.Key(agentID.Bytes())
}

func allAccountsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyAllAccounts)
}

func allAccountsMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyAllAccounts)
}

func accountExists(state kv.KVStoreReader, agentID isc.AgentID, chainID isc.ChainID) bool {
	return allAccountsMapR(state).HasAt([]byte(accountKey(agentID, chainID)))
}

func allAccountsAsDict(state kv.KVStoreReader) dict.Dict {
	ret := dict.New()
	allAccountsMapR(state).IterateKeys(func(agentID []byte) bool {
		ret.Set(kv.Key(agentID), []byte{0x01})
		return true
	})
	return ret
}

// touchAccount ensures the account is in the list of all accounts
func touchAccount(state kv.KVStore, agentID isc.AgentID, chainID isc.ChainID) {
	allAccountsMap(state).SetAt([]byte(accountKey(agentID, chainID)), codec.EncodeBool(true))
}

// HasEnoughForAllowance checks whether an account has enough balance to cover for the allowance
func HasEnoughForAllowance(state kv.KVStoreReader, agentID isc.AgentID, allowance *isc.Assets, chainID isc.ChainID) bool {
	if allowance == nil || allowance.IsEmpty() {
		return true
	}
	accountKey := accountKey(agentID, chainID)
	if allowance != nil {
		if getBaseTokens(state, accountKey) < allowance.BaseTokens {
			return false
		}
		for _, nativeToken := range allowance.NativeTokens {
			if getNativeTokenAmount(state, accountKey, nativeToken.ID).Cmp(nativeToken.Amount) < 0 {
				return false
			}
		}
	}
	for _, nftID := range allowance.NFTs {
		if !hasNFT(state, agentID, nftID) {
			return false
		}
	}
	return true
}

// MoveBetweenAccounts moves assets between on-chain accounts
func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) error {
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return nil
	}

	if !debitFromAccount(state, accountKey(fromAgentID, chainID), assets) {
		return errors.New("MoveBetweenAccounts: not enough funds")
	}
	creditToAccount(state, accountKey(toAgentID, chainID), assets)

	for _, nftID := range assets.NFTs {
		nft := GetNFTData(state, nftID)
		if nft == nil {
			return fmt.Errorf("MoveBetweenAccounts: unknown NFT %s", nftID)
		}
		if !debitNFTFromAccount(state, fromAgentID, nft) {
			return errors.New("MoveBetweenAccounts: NFT not found in origin account")
		}
		creditNFTToAccount(state, toAgentID, nft.ID, nft.Issuer)
	}

	touchAccount(state, fromAgentID, chainID)
	touchAccount(state, toAgentID, chainID)
	return nil
}

func MustMoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) {
	err := MoveBetweenAccounts(state, fromAgentID, toAgentID, assets, chainID)
	if err != nil {
		panic(err)
	}
}

// debitBaseTokensFromAllowance is used for adjustment of L2 when part of base tokens are taken for storage deposit
// It takes base tokens from allowance to the common account and then removes them from the L2 ledger
func debitBaseTokensFromAllowance(ctx isc.Sandbox, amount uint64, chainID isc.ChainID) {
	if amount == 0 {
		return
	}
	storageDepositAssets := isc.NewAssetsBaseTokens(amount)
	ctx.TransferAllowedFunds(CommonAccount(), storageDepositAssets)
	DebitFromAccount(ctx.State(), CommonAccount(), storageDepositAssets, chainID)
}

func UpdateLatestOutputID(state kv.KVStore, anchorTxID iotago.TransactionID, blockIndex uint32) {
	updateNativeTokenOutputIDs(state, anchorTxID)
	updateFoundryOutputIDs(state, anchorTxID)
	updateNFTOutputIDs(state, anchorTxID)
	updateNewlyMintedNFTOutputIDs(state, anchorTxID, blockIndex)
}
