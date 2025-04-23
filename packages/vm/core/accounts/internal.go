package accounts

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var (
	ErrNotEnoughFunds       = coreerrors.Register("not enough funds").Create()
	ErrNotEnoughAllowance   = coreerrors.Register("not enough allowance").Create()
	ErrBadAmount            = coreerrors.Register("bad native asset amount").Create()
	ErrDuplicateTreasuryCap = coreerrors.Register("duplicate TreasuryCap").Create()
	ErrTreasuryCapNotFound  = coreerrors.Register("TreasuryCap not found").Create()
	ErrOverflow             = coreerrors.Register("overflow in token arithmetics").Create()
)

var (
	// KeyAllAccounts stores a map of <agentID> => true
	// Covered in: TestFoundries
	KeyAllAccounts = "a"

	// PrefixAccountCoinBalances | <accountID> stores a map of <coinType> => coin.Value
	// Covered in: TestFoundries
	PrefixAccountCoinBalances = "C"

	// PrefixAccountWeiRemainder | <accountID> stores the wei remainder (big.Int 18 decimals)
	PrefixAccountWeiRemainder kv.Key = "w"

	// L2TotalsAccount is the special <accountID> storing the total coin balances
	// controlled by the chain
	// Covered in: TestFoundries
	L2TotalsAccount kv.Key = "*"

	// keyNonce stores a map of <agentID> => nonce (uint64)
	// Covered in: TODO
	KeyNonce kv.Key = "m"

	// KeyCoinInfo stores a map of <CoinType> => isc.IotaCoinInfo
	// Covered in: TestFoundries
	KeyCoinInfo = "RC"

	// prefixObjects | <agentID> stores a map of <ObjectID> => <ObjectType>
	// Covered in: TODO
	PrefixObjects = "o"

	// keyObjectOwner stores a map of <ObjectID> => isc.AgentID
	// Covered in: TODO
	KeyObjectOwner = "W"
)

func AccountKey(agentID isc.AgentID) kv.Key {
	return kv.Key(agentID.Bytes())
}

func (s *StateWriter) allAccountsMap() *collections.Map {
	return collections.NewMap(s.state, KeyAllAccounts)
}

func (s *StateReader) allAccountsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, KeyAllAccounts)
}

func (s *StateReader) AccountExists(agentID isc.AgentID) bool {
	return s.allAccountsMapR().HasAt([]byte(AccountKey(agentID)))
}

func (s *StateReader) AllAccountsAsDict() dict.Dict {
	ret := dict.New()
	s.allAccountsMapR().IterateKeys(func(accKey []byte) bool {
		ret.Set(kv.Key(accKey), []byte{0x01})
		return true
	})
	return ret
}

func (s *StateReader) IterateAllAccounts(f func(accKey []byte) bool) {
	s.allAccountsMapR().IterateKeys(f)
}

// touchAccount ensures the account is in the list of all accounts
func (s *StateWriter) touchAccount(agentID isc.AgentID) {
	s.allAccountsMap().SetAt([]byte(AccountKey(agentID)), codec.Encode(true))
}

// HasEnoughForAllowance checks whether an account has enough balance to cover for the allowance
func (s *StateReader) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets) bool {
	if allowance == nil || allowance.IsEmpty() {
		return true
	}
	accountKey := AccountKey(agentID)
	for coinType, amount := range allowance.Coins.Iterate() {
		if s.getCoinBalance(accountKey, coinType) < amount {
			return false
		}
	}
	for obj := range allowance.Objects.Iterate() {
		if !s.hasObject(agentID, obj.ID) {
			return false
		}
	}
	return true
}

// MoveBetweenAccounts moves assets between on-chain accounts
func (s *StateWriter) MoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) error {
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return nil
	}

	if !s.debitFromAccount(AccountKey(fromAgentID), assets.Coins) {
		return errors.New("MoveBetweenAccounts: not enough funds")
	}
	s.creditToAccount(AccountKey(toAgentID), assets.Coins)

	for obj := range assets.Objects.Iterate() {
		_, ok := s.removeObjectOwner(obj.ID, fromAgentID)
		if !ok {
			return errors.New("MoveBetweenAccounts: object not found in origin account")
		}
		s.setObjectOwner(obj, toAgentID)
	}

	s.touchAccount(fromAgentID)
	s.touchAccount(toAgentID)
	return nil
}
