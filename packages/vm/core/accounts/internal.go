package accounts

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
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
	ErrInvalidNFTID                         = coreerrors.Register("invalid NFT ID").Create()
	ErrTooManyNFTsInAllowance               = coreerrors.Register("expected at most 1 NFT in allowance").Create()
	ErrNFTIDNotFound                        = coreerrors.Register("NFTID not found: %s")
)

const (
	// keyAllAccounts stores a map of <agentID> => true
	// where sum = baseTokens + native tokens + nfts
	keyAllAccounts = "a"

	// prefixBaseTokens | <accountID> stores the amount of base tokens (big.Int)
	prefixBaseTokens = "b"
	// prefixBaseTokens | <accountID> stores a map of <nativeTokenID> => big.Int
	prefixNativeTokens = "t"

	// l2TotalsAccount is the special <accountID> storing the total fungible tokens
	// controlled by the chain
	l2TotalsAccount = "*"

	// prefixNFTs | <agentID> stores a map of <NFTID> => true
	prefixNFTs = "n"
	// prefixNFTsByCollection | <agentID> | <collectionID> stores a map of <nftID> => true
	prefixNFTsByCollection = "c"
	// prefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	prefixFoundries = "f"

	// noCollection is the special <collectionID> used for storing NFTs that do not belong in a collection
	noCollection = "-"

	// keyMaxAssumedNonce stores a map of <agentID> => max assumed nonce (uint64)
	keyMaxAssumedNonce = "m"

	// keyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	keyNativeTokenOutputMap = "TO"
	// keyFoundryOutputRecords stores a map of <foundrySN> => foundryOutputRec
	keyFoundryOutputRecords = "FO"
	// keyNFTOutputRecords stores a map of <NFTID> => NFTOutputRec
	keyNFTOutputRecords = "NO"
	// keyNFTData stores a map of <NFTID> => isc.NFT
	keyNFTData = "ND"
	// keyStorageDepositAssumptions => transaction.StorageDepositAssumption
	keyStorageDepositAssumptions = "S"
)

func accountKey(agentID isc.AgentID) kv.Key {
	return kv.Key(agentID.Bytes())
}

func allAccountsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyAllAccounts)
}

func allAccountsMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyAllAccounts)
}

func AccountExists(state kv.KVStoreReader, agentID isc.AgentID) bool {
	return allAccountsMapR(state).MustHasAt(agentID.Bytes())
}

func allAccountsAsDict(state kv.KVStoreReader) dict.Dict {
	ret := dict.New()
	allAccountsMapR(state).MustIterate(func(agentID []byte, val []byte) bool {
		ret.Set(kv.Key(agentID), []byte{0xff})
		return true
	})
	return ret
}

// touchAccount ensures the account is in the list of all accounts
func touchAccount(state kv.KVStore, agentID isc.AgentID) {
	allAccountsMap(state).MustSetAt([]byte(accountKey(agentID)), codec.EncodeBool(true))
}

// HasEnoughForAllowance checkes whether an account has enough balance to cover for the allowance
func HasEnoughForAllowance(state kv.KVStoreReader, agentID isc.AgentID, allowance *isc.Allowance) bool {
	if allowance == nil || allowance.IsEmpty() {
		return true
	}
	accountKey := accountKey(agentID)
	if allowance.Assets != nil {
		if getBaseTokens(state, accountKey) < allowance.Assets.BaseTokens {
			return false
		}
		for _, nativeToken := range allowance.Assets.NativeTokens {
			if getNativeTokens(state, accountKey, nativeToken.ID).Cmp(nativeToken.Amount) < 0 {
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
func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, fungibleTokens *isc.FungibleTokens, nfts []iotago.NFTID) error {
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return nil
	}

	if !debitFromAccount(state, accountKey(fromAgentID), fungibleTokens) {
		return errors.New("MoveBetweenAccounts: not enough funds")
	}
	creditToAccount(state, accountKey(toAgentID), fungibleTokens)

	for _, nftID := range nfts {
		nft, err := GetNFTData(state, nftID)
		if err != nil {
			return err
		}
		if !debitNFTFromAccount(state, fromAgentID, nft) {
			return errors.New("MoveBetweenAccounts: NFT not found in origin account")
		}
		creditNFTToAccount(state, toAgentID, nft)
	}

	touchAccount(state, fromAgentID)
	touchAccount(state, toAgentID)
	return nil
}

func MustMoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, fungibleTokens *isc.FungibleTokens, nfts []iotago.NFTID) {
	err := MoveBetweenAccounts(state, fromAgentID, toAgentID, fungibleTokens, nfts)
	if err != nil {
		panic(err)
	}
}

func CheckLedger(state kv.KVStoreReader, checkpoint string) {
	t := GetTotalL2FungibleTokens(state)
	c := calcL2TotalFungibleTokens(state)
	if !t.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n total assets: %s\ncalc total: %s\n",
			checkpoint, t, c))
	}

	totalAccNFTs := GetTotalL2NFTs(state)
	if len(lo.FindDuplicates(totalAccNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	calculatedNFTs := calcL2TotalNFTs(state)
	if len(lo.FindDuplicates(calculatedNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	left, right := lo.Difference(calculatedNFTs, totalAccNFTs)
	if len(left)+len(right) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n NFTs don't match\n", checkpoint))
	}
}

func GetStorageDepositAssumptions(state kv.KVStoreReader) *transaction.StorageDepositAssumption {
	bin := state.MustGet(kv.Key(keyStorageDepositAssumptions))
	ret, err := transaction.StorageDepositAssumptionFromBytes(bin)
	if err != nil {
		panic(fmt.Errorf("GetStorageDepositAssumptions: internal: %w", err))
	}
	return ret
}

// debitBaseTokensFromAllowance is used for adjustment of L2 when part of base tokens are taken for storage deposit
// It takes base tokens from allowance to the common account and then removes them from the L2 ledger
func debitBaseTokensFromAllowance(ctx isc.Sandbox, amount uint64) {
	if amount == 0 {
		return
	}
	commonAccount := ctx.ChainID().CommonAccount()
	storageDepositAssets := isc.NewFungibleBaseTokens(amount)
	transfer := isc.NewAllowanceFungibleTokens(storageDepositAssets)
	ctx.TransferAllowedFunds(commonAccount, transfer)
	DebitFromAccount(ctx.State(), commonAccount, storageDepositAssets)
}
