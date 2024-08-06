package accounts

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/bigint"
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
	ErrDuplicateTreasuryCap                 = coreerrors.Register("duplicate TreasuryCap").Create()
	ErrTreasuryCapNotFound                  = coreerrors.Register("TreasuryCap not found").Create()
	ErrOverflow                             = coreerrors.Register("overflow in token arithmetics").Create()
	ErrTooManyNFTsInAllowance               = coreerrors.Register("expected at most 1 NFT in allowance").Create()
	ErrNFTIDNotFound                        = coreerrors.Register("NFTID not found").Create()
	ErrImmutableMetadataInvalid             = coreerrors.Register("IRC27 metadata is invalid: '%s'")
)

const (
	// keyAllAccounts stores a map of <agentID> => true
	// Covered in: TestFoundries
	keyAllAccounts = "a"

	// prefixCoins | <accountID> stores a map of <CoinType> => big.Int
	// Covered in: TestFoundries
	prefixCoins = "C"

	// prefixRemainders | <accountID> stores the wei remainder (big.Int 18 decimals)
	prefixRemainders = "w"

	// L2TotalsAccount is the special <accountID> storing the total coins
	// controlled by the chain
	// Covered in: TestFoundries
	L2TotalsAccount = "*"

	// prefixObjects | <agentID> stores a map of <ObjectID> => true
	// Covered in: TestDepositNFTWithMinStorageDeposit
	prefixObjects = "o"
	// prefixObjectsByCollection | <agentID> | <collectionID> stores a map of <ObjectID> => true
	// Covered in: TestNFTMint
	// Covered in: TestDepositNFTWithMinStorageDeposit
	prefixObjectsByCollection = "c"

	// prefixTreasuryCaps + <agentID> stores a map of <ObjectID> => true
	// Covered in: TestFoundries
	prefixTreasuryCaps = "T"

	// noCollection is the special <collectionID> used for storing NFTs that do not belong in a collection
	// Covered in: TestNFTMint
	noCollection = "-"

	// keyNonce stores a map of <agentID> => nonce (uint64)
	// Covered in: TestNFTMint
	keyNonce = "m"

	// keyCoinRecords stores a map of <CoinType> => array of CoinRecord
	// Covered in: TestFoundries
	keyCoinRecords = "RC"
	// keyTreasuryCapRecords stores a map of <CoinType> => TreasuryCapRecord
	// Covered in: TestFoundries
	keyTreasuryCapRecords = "RT"
	// keyObjectRecords stores a map of <ObjectID> => ObjectRecord
	// Covered in: TestDepositNFTWithMinStorageDeposit
	keyObjectRecords = "RO"

	// keyObjectOwner stores a map of <ObjectID> => isc.AgentID
	// Covered in: TestDepositNFTWithMinStorageDeposit
	keyObjectOwner = "W"
)

func accountKey(agentID isc.AgentID, chainID isc.ChainID) kv.Key {
	if agentID.BelongsToChain(chainID) {
		// save bytes by skipping the chainID bytes on agentIDs for this chain
		return kv.Key(agentID.BytesWithoutChainID())
	}
	return kv.Key(agentID.Bytes())
}

func (s *StateWriter) allAccountsMap() *collections.Map {
	return collections.NewMap(s.state, keyAllAccounts)
}

func (s *StateReader) allAccountsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyAllAccounts)
}

func (s *StateReader) AccountExists(agentID isc.AgentID, chainID isc.ChainID) bool {
	return s.allAccountsMapR().HasAt([]byte(accountKey(agentID, chainID)))
}

func (s *StateReader) AllAccountsAsDict() dict.Dict {
	ret := dict.New()
	s.allAccountsMapR().IterateKeys(func(accKey []byte) bool {
		ret.Set(kv.Key(accKey), []byte{0x01})
		return true
	})
	return ret
}

// touchAccount ensures the account is in the list of all accounts
func (s *StateWriter) touchAccount(agentID isc.AgentID, chainID isc.ChainID) {
	s.allAccountsMap().SetAt([]byte(accountKey(agentID, chainID)), codec.Bool.Encode(true))
}

// HasEnoughForAllowance checks whether an account has enough balance to cover for the allowance
func (s *StateReader) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets, chainID isc.ChainID) bool {
	if allowance == nil || allowance.IsEmpty() {
		return true
	}
	accountKey := accountKey(agentID, chainID)
	for coinType, amount := range allowance.Coins {
		if bigint.Less(s.getCoinBalance(accountKey, coinType), amount) {
			return false
		}
	}
	for id := range allowance.Objects {
		if !s.hasObject(agentID, id) {
			return false
		}
	}
	return true
}

// MoveBetweenAccounts moves assets between on-chain accounts
func (s *StateWriter) MoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) error {
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return nil
	}

	if !s.debitFromAccount(accountKey(fromAgentID, chainID), assets.Coins) {
		return errors.New("MoveBetweenAccounts: not enough funds")
	}
	s.creditToAccount(accountKey(toAgentID, chainID), assets.Coins)

	for id := range assets.Objects {
		obj := s.GetObject(id)
		if obj == nil {
			return fmt.Errorf("MoveBetweenAccounts: unknown object %s", id)
		}
		if !s.debitObjectFromAccount(fromAgentID, obj) {
			return errors.New("MoveBetweenAccounts: NFT not found in origin account")
		}
		s.creditObjectToAccount(toAgentID, obj)
	}

	s.touchAccount(fromAgentID, chainID)
	s.touchAccount(toAgentID, chainID)
	return nil
}

// debitBaseTokensFromAllowance is used for adjustment of L2 when part of base tokens are taken for storage deposit
// It takes base tokens from allowance to the common account and then removes them from the L2 ledger
func debitBaseTokensFromAllowance(ctx isc.Sandbox, amount uint64, chainID isc.ChainID) {
	if amount == 0 {
		return
	}
	storageDepositAssets := isc.NewAssetsBaseTokens(amount)
	ctx.TransferAllowedFunds(CommonAccount(), storageDepositAssets)
	NewStateWriterFromSandbox(ctx).DebitFromAccount(CommonAccount(), storageDepositAssets.Coins, chainID)
}
