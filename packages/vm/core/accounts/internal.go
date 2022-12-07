package accounts

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
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

// getAccount each account is a map with the name of its controlling agentID.
// - nil key is balance of base tokens uint64 8 bytes little-endian
// - iotago.NativeTokenID key is a big.Int balance of the native token
func getAccount(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, string(kv.Concat(prefixAccount, agentID.Bytes())))
}

func getAccountR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, string(kv.Concat(prefixAccount, agentID.Bytes())))
}

// getTotalL2AssetsAccount is an account with totals by token type
func getTotalL2AssetsAccount(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixTotalL2AssetsAccount)
}

func getTotalL2AssetsAccountR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixTotalL2AssetsAccount)
}

// GetAccountsMap is a map which contains all non-empty accounts
func GetAccountsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixAllAccounts)
}

func getAccountsMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixAllAccounts)
}

func nonceKey(callerAgentID isc.AgentID) kv.Key {
	return kv.Key(kv.Concat(prefixMaxAssumedNonceKey, callerAgentID.Bytes()))
}

func getAccountFoundries(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, string(kv.Concat(prefixAccountFoundries, agentID.Bytes())))
}

func getAccountFoundriesR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, string(kv.Concat(prefixAccountFoundries, agentID.Bytes())))
}

func AccountExists(state kv.KVStoreReader, agentID isc.AgentID) bool {
	return getAccountR(state, agentID).MustLen() > 0
}

// TODO getNFTState -> getNFTDirectory ???
func getNFTState(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixNFTData)
}

func getNFTStateR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixNFTData)
}

// GetMaxAssumedNonce is maintained for each caller with the purpose of replay protection of off-ledger requests
func GetMaxAssumedNonce(state kv.KVStoreReader, callerAgentID isc.AgentID) uint64 {
	nonce, err := codec.DecodeUint64(state.MustGet(nonceKey(callerAgentID)), 0)
	if err != nil {
		panic(xerrors.Errorf("GetMaxAssumedNonce: %w", err))
	}
	return nonce
}

func SaveMaxAssumedNonce(state kv.KVStore, callerAgentID isc.AgentID, nonce uint64) {
	next := GetMaxAssumedNonce(state, callerAgentID) + 1
	if nonce > next {
		next = nonce
	}
	state.Set(nonceKey(callerAgentID), codec.EncodeUint64(next))
}

// touchAccount ensures that only non-empty accounts are kept in the accounts map
func touchAccount(state kv.KVStore, account *collections.Map) {
	if account.Name() == prefixTotalL2AssetsAccount {
		return
	}
	agentid := []byte(account.Name())[1:] // skip the prefix
	accounts := GetAccountsMap(state)
	if account.MustLen() == 0 {
		accounts.MustDelAt(agentid)
	} else {
		accounts.MustSetAt(agentid, []byte{0xFF})
	}
}

// tokenBalanceMutation structure for handling mutations of the on-chain accounts
type tokenBalanceMutation struct {
	balance *big.Int
	delta   *big.Int
}

// loadAccountMutations traverses the assets of interest in the account and collects values for further processing
func loadAccountMutations(account *collections.Map, assets *isc.FungibleTokens) (uint64, uint64, map[iotago.NativeTokenID]tokenBalanceMutation) {
	if assets == nil {
		return 0, 0, nil
	}

	addBaseTokens := assets.BaseTokens
	fromBaseTokens := uint64(0)
	if v := account.MustGetAt(nil); v != nil {
		fromBaseTokens = util.MustUint64From8Bytes(v)
	}

	tokenMutations := make(map[iotago.NativeTokenID]tokenBalanceMutation)
	for _, nt := range assets.Tokens {
		if nt.Amount.Cmp(util.Big0) < 0 {
			panic(ErrBadAmount)
		}
		bal := big.NewInt(0)
		if v := account.MustGetAt(nt.ID[:]); v != nil {
			bal.SetBytes(v)
		}
		tokenMutations[nt.ID] = tokenBalanceMutation{
			balance: bal,
			delta:   nt.Amount,
		}
	}
	return fromBaseTokens, addBaseTokens, tokenMutations
}

// CreditToAccount brings new funds to the on chain ledger
func CreditToAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.FungibleTokens) {
	if assets == nil || (assets.BaseTokens == 0 && len(assets.Tokens) == 0) {
		return
	}
	account := getAccount(state, agentID)

	creditToAccount(account, assets)
	creditToAccount(getTotalL2AssetsAccount(state), assets)
	touchAccount(state, account)
}

// creditToAccount adds assets to the internal account map
func creditToAccount(account *collections.Map, assets *isc.FungibleTokens) {
	balance, add, tokenMutations := loadAccountMutations(account, assets)
	if add > 0 {
		account.MustSetAt(nil, util.Uint64To8Bytes(balance+add))
	}
	for assetID, m := range tokenMutations {
		if util.IsZeroBigInt(m.delta) {
			continue
		}
		// safe arithmetics
		if m.delta.Cmp(big.NewInt(0)) < 0 {
			panic(ErrBadAmount)
		}
		if m.balance.Cmp(new(big.Int).Sub(abi.MaxUint256, m.delta)) > 0 {
			panic(ErrOverflow)
		}
		m.balance.Add(m.balance, m.delta)
		account.MustSetAt(assetID[:], m.balance.Bytes())
	}
}

// DebitFromAccount takes out assets balance the on chain ledger. If not enough it panics
func DebitFromAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.FungibleTokens) {
	if assets.IsEmpty() {
		return
	}
	account := getAccount(state, agentID)

	if !debitFromAccount(account, assets) {
		panic(xerrors.Errorf("debit from %s: %v\nassets: %s", agentID, ErrNotEnoughFunds, assets))
	}
	if !debitFromAccount(getTotalL2AssetsAccount(state), assets) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	touchAccount(state, account)
}

// debitFromAccount debits assets from the internal accounts map
func debitFromAccount(account *collections.Map, assets *isc.FungibleTokens) bool {
	balance, sub, tokenMutations := loadAccountMutations(account, assets)
	// check if enough
	if balance < sub {
		return false
	}
	for _, m := range tokenMutations {
		if m.balance.Cmp(m.delta) < 0 {
			return false
		}
	}
	if sub > 0 {
		if balance == sub {
			account.MustDelAt(nil)
		} else {
			account.MustSetAt(nil, util.Uint64To8Bytes(balance-sub))
		}
	}
	for id, m := range tokenMutations {
		m.balance.Sub(m.balance, m.delta)
		if util.IsZeroBigInt(m.balance) {
			account.MustDelAt(id[:])
		} else {
			account.MustSetAt(id[:], m.balance.Bytes())
		}
	}
	return true
}

// GetNativeTokenBalance returns balance or nil if it does not exist
func HasEnoughForAllowance(state kv.KVStoreReader, agentID isc.AgentID, allowance *isc.Allowance) bool {
	return hasEnoughForAllowance(getAccountR(state, agentID), allowance)
}

// enoughForAllowance checkes whether an account has enough balance to cover for the allowance
func hasEnoughForAllowance(account *collections.ImmutableMap, allowance *isc.Allowance) bool {
	if allowance == nil || allowance.IsEmpty() {
		return true
	}
	if account.MustLen() == 0 {
		return false
	}
	// check base token
	if allowance.Assets != nil {
		accountBaseTokenS := util.MustUint64From8Bytes(account.MustGetAt(nil))
		if accountBaseTokenS < allowance.Assets.BaseTokens {
			return false
		}

		// check native tokens
		for _, token := range allowance.Assets.Tokens {
			v := account.MustGetAt(token.ID[:])
			if v == nil {
				return false
			}
			bal := big.NewInt(0)
			bal.SetBytes(v)
			if bal.Cmp(token.Amount) == -1 {
				return false
			}
		}
	}

	// check NFTs
	for _, nftID := range allowance.NFTs {
		v := account.MustGetAt(nftID[:])
		if v == nil {
			return false
		}
	}

	return true
}

// MoveBetweenAccounts moves assets between on-chain accounts. Returns if it was a success (= enough funds in the source)
func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, fungibleTokens *isc.FungibleTokens, nfts []iotago.NFTID) bool {
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return true
	}
	// total assets doesn't change
	fromAccount := getAccount(state, fromAgentID)
	toAccount := getAccount(state, toAgentID)
	if !debitFromAccount(fromAccount, fungibleTokens) {
		return false
	}
	creditToAccount(toAccount, fungibleTokens)

	defer func() {
		touchAccount(state, fromAccount)
		touchAccount(state, toAccount)
	}()

	for _, nft := range nfts {
		if !debitNFTFromAccount(fromAccount, nft) {
			return false
		}
		creditNFTToAccount(state, toAccount, nft, toAgentID)
	}

	return true
}

func MustMoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID isc.AgentID, fungibleTokens *isc.FungibleTokens, nfts []iotago.NFTID) {
	if !MoveBetweenAccounts(state, fromAgentID, toAgentID, fungibleTokens, nfts) {
		panic(xerrors.Errorf(" agentID: %s. %v. fungibleTokens: %s, nfts: %s", fromAgentID, ErrNotEnoughFunds, fungibleTokens, nfts))
	}
}

func AdjustAccountBaseTokens(state kv.KVStore, account isc.AgentID, adjustment int64) {
	switch {
	case adjustment > 0:
		CreditToAccount(state, account, isc.NewFungibleTokens(uint64(adjustment), nil))
	case adjustment < 0:
		DebitFromAccount(state, account, isc.NewFungibleTokens(uint64(-adjustment), nil))
	}
}

// GetBaseTokensBalance return base tokens balance. 0 means it does not exist
func GetBaseTokensBalance(state kv.KVStoreReader, agentID isc.AgentID) uint64 {
	return getBaseTokensBalance(getAccountR(state, agentID))
}

func getBaseTokensBalance(account *collections.ImmutableMap) uint64 {
	if v := account.MustGetAt(nil); v != nil {
		return util.MustUint64From8Bytes(v)
	}
	return 0
}

// GetNativeTokenBalance returns balance or nil if it does not exist
func GetNativeTokenBalance(state kv.KVStoreReader, agentID isc.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	return getNativeTokenBalance(getAccountR(state, agentID), tokenID)
}

func GetNativeTokenBalanceTotal(state kv.KVStoreReader, tokenID *iotago.NativeTokenID) *big.Int {
	return getNativeTokenBalance(getTotalL2AssetsAccountR(state), tokenID)
}

func getNativeTokenBalance(account *collections.ImmutableMap, tokenID *iotago.NativeTokenID) *big.Int {
	ret := big.NewInt(0)
	if v := account.MustGetAt(tokenID[:]); v != nil {
		return ret.SetBytes(v)
	}
	return ret
}

func getAccountsIntern(state kv.KVStoreReader) dict.Dict {
	ret := dict.New()
	getAccountsMapR(state).MustIterate(func(agentID []byte, val []byte) bool {
		ret.Set(kv.Key(agentID), []byte{0xff})
		return true
	})
	return ret
}

func getAccountAssets(account *collections.ImmutableMap) *isc.FungibleTokens {
	ret := isc.NewEmptyFungibleTokens()
	account.MustIterate(func(idBytes []byte, val []byte) bool {
		if len(idBytes) == 0 {
			ret.BaseTokens = util.MustUint64From8Bytes(val)
			return true
		}
		if len(idBytes) != iotago.NativeTokenIDLength {
			return true // NFT or some other asset that is not a native token
		}
		token := iotago.NativeToken{
			ID:     isc.MustNativeTokenIDFromBytes(idBytes),
			Amount: new(big.Int).SetBytes(val),
		}
		ret.Tokens = append(ret.Tokens, &token)
		return true
	})
	return ret
}

// GetAccountAssets returns all assets belonging to the agentID on the state
func GetAccountAssets(state kv.KVStoreReader, agentID isc.AgentID) *isc.FungibleTokens {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return isc.NewEmptyFungibleTokens()
	}
	return getAccountAssets(account)
}

func GetTotalL2Assets(state kv.KVStoreReader) *isc.FungibleTokens {
	return getAccountAssets(getTotalL2AssetsAccountR(state))
}

// calcL2TotalAssets traverses the ledger and sums up all assets
func calcL2TotalAssets(state kv.KVStoreReader) *isc.FungibleTokens {
	ret := isc.NewEmptyFungibleTokens()

	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := isc.AgentIDFromBytes(key)
		if err != nil {
			panic(xerrors.Errorf("calcL2TotalAssets: %w", err))
		}
		accBalances := getAccountAssets(getAccountR(state, agentID))
		ret.Add(accBalances)
		return true
	})
	return ret
}

// GetAccountNFTs returns all NFTs belonging to the agentID on the state
func GetAccountNFTs(state kv.KVStoreReader, agentID isc.AgentID) []iotago.NFTID {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil
	}
	return getAccountNFTs(account)
}

func getAccountNFTs(account *collections.ImmutableMap) []iotago.NFTID {
	ret := []iotago.NFTID{}
	account.MustIterate(func(idBytes []byte, val []byte) bool {
		if len(idBytes) != iotago.NFTIDLength {
			return true // Native token or some other asset that is not an NFT
		}
		id := iotago.NFTID{}
		copy(id[:], idBytes)
		ret = append(ret, id)
		return true
	})
	return ret
}

func GetTotalL2NFTs(state kv.KVStoreReader) map[iotago.NFTID]bool {
	ret := make(map[iotago.NFTID]bool)
	nftMap := getNFTStateR(state)
	nftMap.MustIterateKeys(func(key []byte) bool {
		id := iotago.NFTID{}
		copy(id[:], key)
		ret[id] = true
		return true
	})
	return ret
}

func calcL2TotalNFTs(state kv.KVStoreReader) map[iotago.NFTID]bool {
	ret := make(map[iotago.NFTID]bool)
	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := isc.AgentIDFromBytes(key)
		if err != nil {
			panic(xerrors.Errorf("calcL2TotalAssets: %w", err))
		}
		accNFTs := getAccountNFTs(getAccountR(state, agentID))
		for _, nft := range accNFTs {
			if ret[nft] {
				panic(fmt.Sprintf("inconsistency: NFT %s is owned by more than 1 account", nft.String()))
			}
			ret[nft] = true
		}
		return true
	})
	return ret
}

func NFTMapEqual(a, b map[iotago.NFTID]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for NFTId := range a {
		if !b[NFTId] {
			return false
		}
	}
	return true
}

func CheckLedger(state kv.KVStoreReader, checkpoint string) {
	a := GetTotalL2Assets(state)
	c := calcL2TotalAssets(state)
	if !a.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n total assets: %s\ncalc total: %s\n",
			checkpoint, a.String(), c.String()))
	}

	totalAccNFTs := GetTotalL2NFTs(state)
	calculatedNFTs := calcL2TotalNFTs(state)
	if !NFTMapEqual(totalAccNFTs, calculatedNFTs) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n NFTs don't match\n", checkpoint))
	}
}

func getAccountBalanceDict(account *collections.ImmutableMap) dict.Dict {
	return getAccountAssets(account).ToDict()
}

// region Foundry outputs ////////////////////////////////////////

// foundryOutputRec contains information to reconstruct output
type foundryOutputRec struct {
	Amount      uint64 // always storage deposit
	TokenScheme iotago.TokenScheme
	Metadata    []byte
	BlockIndex  uint32
	OutputIndex uint16
}

func (f *foundryOutputRec) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(f.BlockIndex).
		WriteUint16(f.OutputIndex).
		WriteUint64(f.Amount)
	util.WriteBytes8ToMarshalUtil(codec.EncodeTokenScheme(f.TokenScheme), mu)
	util.WriteBytes16ToMarshalUtil(f.Metadata, mu)

	return mu.Bytes()
}

func foundryOutputRecFromMarshalUtil(mu *marshalutil.MarshalUtil) (*foundryOutputRec, error) {
	ret := &foundryOutputRec{}
	var err error
	if ret.BlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.OutputIndex, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if ret.Amount, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	schemeBin, err := util.ReadBytes8FromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	if ret.TokenScheme, err = codec.DecodeTokenScheme(schemeBin); err != nil {
		return nil, err
	}
	if ret.Metadata, err = util.ReadBytes16FromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func mustFoundryOutputRecFromBytes(data []byte) *foundryOutputRec {
	ret, err := foundryOutputRecFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		panic(err)
	}
	return ret
}

// GetAccountsMap is a map which contains all foundries owned by the chain
func getFoundriesMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixFoundryOutputRecords)
}

func getFoundriesMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixFoundryOutputRecords)
}

// SaveFoundryOutput stores foundry output into the map of all foundry outputs (compressed form)
func SaveFoundryOutput(state kv.KVStore, f *iotago.FoundryOutput, blockIndex uint32, outputIndex uint16) {
	foundryRec := foundryOutputRec{
		Amount:      f.Amount,
		TokenScheme: f.TokenScheme,
		Metadata:    []byte{},
		BlockIndex:  blockIndex,
		OutputIndex: outputIndex,
	}
	getFoundriesMap(state).MustSetAt(util.Uint32To4Bytes(f.SerialNumber), foundryRec.Bytes())
}

// DeleteFoundryOutput deletes foundry output from the map of all foundries
func DeleteFoundryOutput(state kv.KVStore, sn uint32) {
	getFoundriesMap(state).MustDelAt(util.Uint32To4Bytes(sn))
}

// GetFoundryOutput returns foundry output, its block number and output index
func GetFoundryOutput(state kv.KVStoreReader, sn uint32, chainID isc.ChainID) (*iotago.FoundryOutput, uint32, uint16) {
	data := getFoundriesMapR(state).MustGetAt(util.Uint32To4Bytes(sn))
	if data == nil {
		return nil, 0, 0
	}
	rec := mustFoundryOutputRecFromBytes(data)

	ret := &iotago.FoundryOutput{
		Amount:       rec.Amount,
		NativeTokens: nil,
		SerialNumber: sn,
		TokenScheme:  rec.TokenScheme,
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: chainID.AsAddress().(*iotago.AliasAddress)},
		},
		Features: nil,
	}
	return ret, rec.BlockIndex, rec.OutputIndex
}

// AddFoundryToAccount ads new foundry to the foundries controlled by the account
func AddFoundryToAccount(state kv.KVStore, agentID isc.AgentID, sn uint32) {
	assert.NewAssert(nil, "check foundry exists")
	addFoundryToAccount(getAccountFoundries(state, agentID), sn)
}

func addFoundryToAccount(account *collections.Map, sn uint32) {
	key := util.Uint32To4Bytes(sn)
	if account.MustHasAt(key) {
		panic(ErrRepeatingFoundrySerialNumber)
	}
	account.MustSetAt(key, []byte{0xFF})
}

func deleteFoundryFromAccount(account *collections.Map, sn uint32) {
	key := util.Uint32To4Bytes(sn)
	if !hasFoundry(account.Immutable(), sn) {
		panic(ErrFoundryNotFound)
	}
	account.MustDelAt(key)
}

// MoveFoundryBetweenAccounts changes ownership of the foundry
func MoveFoundryBetweenAccounts(state kv.KVStore, agentIDFrom, agentIDTo isc.AgentID, sn uint32) {
	deleteFoundryFromAccount(getAccountFoundries(state, agentIDFrom), sn)
	addFoundryToAccount(getAccountFoundries(state, agentIDTo), sn)
}

// HasFoundry checks if specific account owns the foundry
func HasFoundry(state kv.KVStoreReader, agentID isc.AgentID, sn uint32) bool {
	return hasFoundry(getAccountFoundriesR(state, agentID), sn)
}

func hasFoundry(account *collections.ImmutableMap, sn uint32) bool {
	return account.MustHasAt(util.Uint32To4Bytes(sn))
}

// endregion ///////////////////////////////////////////////////////////////////

// region NativeToken outputs /////////////////////////////////

type nativeTokenOutputRec struct {
	StorageBaseTokens uint64 // always storage deposit
	Amount            *big.Int
	BlockIndex        uint32
	OutputIndex       uint16
}

func (f *nativeTokenOutputRec) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(f.BlockIndex).
		WriteUint16(f.OutputIndex).
		WriteUint64(f.StorageBaseTokens)
	util.WriteBytes8ToMarshalUtil(codec.EncodeBigIntAbs(f.Amount), mu)
	return mu.Bytes()
}

func (f *nativeTokenOutputRec) String() string {
	return fmt.Sprintf("Native Token Account: base tokens: %d, amount: %d, block: %d, outIdx: %d",
		f.StorageBaseTokens, f.Amount, f.BlockIndex, f.OutputIndex)
}

func nativeTokenOutputRecFromMarshalUtil(mu *marshalutil.MarshalUtil) (*nativeTokenOutputRec, error) {
	ret := &nativeTokenOutputRec{}
	var err error
	if ret.BlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.OutputIndex, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if ret.StorageBaseTokens, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	bigIntBin, err := util.ReadBytes8FromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.Amount = big.NewInt(0).SetBytes(bigIntBin)
	return ret, nil
}

func mustNativeTokenOutputRecFromBytes(data []byte) *nativeTokenOutputRec {
	ret, err := nativeTokenOutputRecFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		panic(err)
	}
	return ret
}

// GetAccountsMap is a map which contains all foundries owned by the chain
func getNativeTokenOutputMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixNativeTokenOutputMap)
}

func getNativeTokenOutputMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixNativeTokenOutputMap)
}

// SaveNativeTokenOutput map tokenID -> foundryRec
func SaveNativeTokenOutput(state kv.KVStore, out *iotago.BasicOutput, blockIndex uint32, outputIndex uint16) {
	tokenRec := nativeTokenOutputRec{
		StorageBaseTokens: out.Amount,
		Amount:            out.NativeTokens[0].Amount,
		BlockIndex:        blockIndex,
		OutputIndex:       outputIndex,
	}
	getNativeTokenOutputMap(state).MustSetAt(out.NativeTokens[0].ID[:], tokenRec.Bytes())
}

func DeleteNativeTokenOutput(state kv.KVStore, tokenID iotago.NativeTokenID) {
	getNativeTokenOutputMap(state).MustDelAt(tokenID[:])
}

func GetNativeTokenOutput(state kv.KVStoreReader, tokenID *iotago.NativeTokenID, chainID isc.ChainID) (*iotago.BasicOutput, uint32, uint16) {
	data := getNativeTokenOutputMapR(state).MustGetAt(tokenID[:])
	if data == nil {
		return nil, 0, 0
	}
	tokenRec := mustNativeTokenOutputRecFromBytes(data)
	ret := &iotago.BasicOutput{
		Amount: tokenRec.StorageBaseTokens,
		NativeTokens: iotago.NativeTokens{{
			ID:     *tokenID,
			Amount: tokenRec.Amount,
		}},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: chainID.AsAddress()},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: chainID.AsAddress(),
			},
		},
	}
	return ret, tokenRec.BlockIndex, tokenRec.OutputIndex
}

// endregion //////////////////////////////////////////

// region NFT outputs /////////////////////////////////
type NFTOutputRec struct {
	Output      *iotago.NFTOutput
	BlockIndex  uint32
	OutputIndex uint16
}

func (r *NFTOutputRec) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(r.BlockIndex).
		WriteUint16(r.OutputIndex)
	outBytes, err := r.Output.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic("error serializing NFToutput")
	}
	mu.WriteBytes(outBytes)
	return mu.Bytes()
}

func (r *NFTOutputRec) String() string {
	return fmt.Sprintf("NFT Record: base tokens: %d, ID: %s, block: %d, outIdx: %d",
		r.Output.Deposit(), r.Output.NFTID, r.BlockIndex, r.OutputIndex)
}

func NFTOutputRecFromMarshalUtil(mu *marshalutil.MarshalUtil) (*NFTOutputRec, error) {
	ret := &NFTOutputRec{}
	var err error
	if ret.BlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if ret.OutputIndex, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	ret.Output = &iotago.NFTOutput{}
	if _, err := ret.Output.Deserialize(mu.ReadRemainingBytes(), serializer.DeSeriModeNoValidation, nil); err != nil {
		return nil, err
	}
	return ret, nil
}

func mustNFTOutputRecFromBytes(data []byte) *NFTOutputRec {
	ret, err := NFTOutputRecFromMarshalUtil(marshalutil.New(data))
	if err != nil {
		panic(err)
	}
	return ret
}

// GetAccountsMap is a map which contains all foundries owned by the chain
func getNFTOutputMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixNFTOutputRecords)
}

func getNFTOutputMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixNFTOutputRecords)
}

// SaveNFTOutput map tokenID -> foundryRec
func SaveNFTOutput(state kv.KVStore, out *iotago.NFTOutput, blockIndex uint32, outputIndex uint16) {
	tokenRec := NFTOutputRec{
		Output:      out,
		BlockIndex:  blockIndex,
		OutputIndex: outputIndex,
	}
	getNFTOutputMap(state).MustSetAt(out.NFTID[:], tokenRec.Bytes())
}

func DeleteNFTOutput(state kv.KVStore, id iotago.NFTID) {
	getNFTOutputMap(state).MustDelAt(id[:])
}

func GetNFTOutput(state kv.KVStoreReader, id iotago.NFTID, chainID isc.ChainID) (*iotago.NFTOutput, uint32, uint16) {
	data := getNFTOutputMapR(state).MustGetAt(id[:])
	if data == nil {
		return nil, 0, 0
	}
	tokenRec := mustNFTOutputRecFromBytes(data)
	return tokenRec.Output, tokenRec.BlockIndex, tokenRec.OutputIndex
}

// endregion //////////////////////////////////////////

func GetStorageDepositAssumptions(state kv.KVStoreReader) *transaction.StorageDepositAssumption {
	bin := state.MustGet(kv.Key(stateVarMinimumStorageDepositAssumptionsBin))
	ret, err := transaction.StorageDepositAssumptionFromBytes(bin)
	if err != nil {
		panic(xerrors.Errorf("GetStorageDepositAssumptions: internal: %v", err))
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
