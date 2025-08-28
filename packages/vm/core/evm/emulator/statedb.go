// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

const (
	keyAccountNonce          = "n" // covered in: TestStorageContract
	keyAccountCode           = "c" // covered in: TestStorageContract
	keyAccountState          = "s" // covered in: TestStorageContract
	keyAccountSelfDestructed = "S" // covered in: TestSelfDestruct
)

func accountKey(prefix kv.Key, addr common.Address) kv.Key {
	return prefix + kv.Key(addr.Bytes())
}

func accountNonceKey(addr common.Address) kv.Key {
	return accountKey(keyAccountNonce, addr)
}

func accountCodeKey(addr common.Address) kv.Key {
	return accountKey(keyAccountCode, addr)
}

func accountStateKey(addr common.Address, hash common.Hash) kv.Key {
	return accountKey(keyAccountState, addr) + kv.Key(hash[:])
}

func accountSelfDestructedKey(addr common.Address) kv.Key {
	return accountKey(keyAccountSelfDestructed, addr)
}

// StateDB implements vm.StateDB with a kv.KVStore as backend.
// The Ethereum account balance is tied to the L1 balance.
type StateDB struct {
	ctx              Context
	kv               kv.KVStore // subrealm of ctx.State()
	logs             []*types.Log
	snapshots        map[int][]*types.Log
	refund           uint64
	transientStorage transientStorage        // EIP-1153
	newContracts     map[common.Address]bool // EIP-6780
	// originalStorage keeps the pre-transaction value of a storage slot.
	// It is populated on the first write to a slot during a transaction and
	// cleared at the beginning of each transaction in Prepare.
	originalStorage map[common.Address]map[common.Hash]common.Hash
}

var _ vm.StateDB = &StateDB{}

func NewStateDB(ctx Context) *StateDB {
	return &StateDB{
		ctx:              ctx,
		kv:               StateDBSubrealm(ctx.State()),
		snapshots:        make(map[int][]*types.Log),
		transientStorage: newTransientStorage(),
		newContracts:     make(map[common.Address]bool),
		originalStorage:  make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func CreateAccount(kv kv.KVStore, addr common.Address) {
	SetNonce(kv, addr, 0)
}

func (s *StateDB) CreateAccount(addr common.Address) {
	CreateAccount(s.kv, addr)
}

func (s *StateDB) GetStateAndCommittedState(addr common.Address, hash common.Hash) (common.Hash, common.Hash) {
	current := s.GetState(addr, hash)
	committed := s.GetCommittedState(addr, hash)
	return current, committed
}

// CreateContract is used whenever a contract is created. This may be preceded
// by CreateAccount, but that is not required if it already existed in the
// state due to funds sent beforehand.
// This operation sets the 'newContract'-flag, which is required in order to
// correctly handle EIP-6780 'delete-in-same-transaction' logic.
func (s *StateDB) CreateContract(addr common.Address) {
	s.CreateAccount(addr)
	s.newContracts[addr] = true
}

// GetStorageRoot implements vm.StateDB.
func (s *StateDB) GetStorageRoot(addr common.Address) common.Hash {
	return common.BytesToHash([]byte(accountStateKey(addr, common.Hash{})))
}

// PointCache implements vm.StateDB.
func (s *StateDB) PointCache() *utils.PointCache {
	panic("unimplemented")
}

func (s *StateDB) SubBalance(addr common.Address, amount *uint256.Int, _ tracing.BalanceChangeReason) uint256.Int {
	prev := uint256.MustFromBig(s.ctx.GetBaseTokensBalance(addr))
	if amount.Sign() == 0 {
		return *prev
	}
	if amount.Sign() == -1 {
		panic("unexpected negative amount")
	}
	s.ctx.SubBaseTokensBalance(addr, amount.ToBig())
	return *prev
}

func (s *StateDB) AddBalance(addr common.Address, amount *uint256.Int, _ tracing.BalanceChangeReason) uint256.Int {
	prev := uint256.MustFromBig(s.ctx.GetBaseTokensBalance(addr))
	if amount.Sign() == 0 {
		return *prev
	}
	if amount.Sign() == -1 {
		panic("unexpected negative amount")
	}
	s.ctx.AddBaseTokensBalance(addr, amount.ToBig())
	return *prev
}

func (s *StateDB) GetBalance(addr common.Address) *uint256.Int {
	return uint256.MustFromBig(s.ctx.GetBaseTokensBalance(addr))
}

func GetNonce(s kv.KVStoreReader, addr common.Address) uint64 {
	return codec.MustDecode[uint64](s.Get(accountNonceKey(addr)), 0)
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	nonce := uint64(0)
	// do not charge gas for this, internal checks of the emulator require this function to run before executing the request
	s.ctx.WithoutGasBurn(func() {
		nonce = GetNonce(s.kv, addr)
	})
	return nonce
}

func IncNonce(kv kv.KVStore, addr common.Address) {
	SetNonce(kv, addr, GetNonce(kv, addr)+1)
}

func (s *StateDB) IncNonce(addr common.Address) {
	IncNonce(s.kv, addr)
}

func SetNonce(kv kv.KVStore, addr common.Address, n uint64) {
	kv.Set(accountNonceKey(addr), codec.Encode(n))
}

func (s *StateDB) SetNonce(addr common.Address, n uint64, r tracing.NonceChangeReason) {
	SetNonce(s.kv, addr, n)
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	return crypto.Keccak256Hash(s.GetCode(addr))
}

func GetCode(s kv.KVStoreReader, addr common.Address) []byte {
	return s.Get(accountCodeKey(addr))
}

func (s *StateDB) GetCode(addr common.Address) []byte {
	return GetCode(s.kv, addr)
}

func SetCode(kv kv.KVStore, addr common.Address, code []byte) {
	if code == nil {
		kv.Del(accountCodeKey(addr))
	} else {
		kv.Set(accountCodeKey(addr), code)
	}
}

func (s *StateDB) SetCode(addr common.Address, code []byte) []byte {
	prev := s.GetCode(addr)
	SetCode(s.kv, addr, code)
	return prev
}

func (s *StateDB) GetCodeSize(addr common.Address) int {
	return len(s.GetCode(addr))
}

func (s *StateDB) AddRefund(n uint64) {
	s.refund += n
}

func (s *StateDB) SubRefund(n uint64) {
	if n > s.refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", n, s.refund))
	}
	s.refund -= n
}

func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

func (s *StateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	if slots, ok := s.originalStorage[addr]; ok {
		if orig, ok2 := slots[key]; ok2 {
			return orig
		}
	}
	return s.GetState(addr, key)
}

func GetState(s kv.KVStoreReader, addr common.Address, key common.Hash) common.Hash {
	return common.BytesToHash(s.Get(accountStateKey(addr, key)))
}

func (s *StateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return GetState(s.kv, addr, key)
}

func SetState(kv kv.KVStore, addr common.Address, key, value common.Hash) {
	kv.Set(accountStateKey(addr, key), value.Bytes())
}

func (s *StateDB) SetState(addr common.Address, key, value common.Hash) common.Hash {
	prev := s.GetState(addr, key)
	// Capture original value on first mutation in this transaction
	if slots, ok := s.originalStorage[addr]; ok {
		if _, exists := slots[key]; !exists {
			// initialize with prev (pre-write value)
			slots[key] = prev
		}
	} else {
		s.originalStorage[addr] = map[common.Hash]common.Hash{key: prev}
	}
	if prev == value {
		return prev
	}
	SetState(s.kv, addr, key, value)
	return prev
}

func (s *StateDB) SelfDestruct(addr common.Address) uint256.Int {
	if !s.Exist(addr) {
		return *uint256.NewInt(0)
	}

	s.kv.Del(accountNonceKey(addr))
	s.kv.Del(accountCodeKey(addr))

	keys := make([]kv.Key, 0)
	s.kv.IterateKeys(accountKey(keyAccountState, addr), func(key kv.Key) bool {
		keys = append(keys, key)
		return true
	})
	for _, k := range keys {
		s.kv.Del(k)
	}

	// for some reason the EVM engine calls AddBalance to the beneficiary address,
	// but not SubBalance for the self-destructed address.
	prevBalance := s.ctx.GetBaseTokensBalance(addr)
	if prevBalance.Sign() > 0 {
		s.ctx.SubBaseTokensBalance(addr, prevBalance)
	}

	s.kv.Set(accountSelfDestructedKey(addr), []byte{1})

	return *uint256.MustFromBig(prevBalance)
}

func (s *StateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	// only allow selfdestruct if within the creation tx (as per EIP-6780)
	if s.newContracts[addr] {
		return s.SelfDestruct(addr), true
	}
	prevBalance := s.ctx.GetBaseTokensBalance(addr)
	return *uint256.MustFromBig(prevBalance), false
}

func (s *StateDB) HasSelfDestructed(addr common.Address) bool {
	return s.kv.Has(accountSelfDestructedKey(addr))
}

// Exist reports whether the given account exists in state.
// Notably this should also return true for self-destructed accounts.
func (s *StateDB) Exist(addr common.Address) bool {
	return Exist(addr, s.kv)
}

// Exist reports whether the given account exists in state.
// expects s to be the stateDB state partition
func Exist(addr common.Address, s kv.KVStoreReader) bool {
	return s.Has(accountNonceKey(addr))
}

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (s *StateDB) Empty(addr common.Address) bool {
	return s.GetNonce(addr) == 0 && s.GetBalance(addr).Sign() == 0 && s.GetCodeSize(addr) == 0
}

func (s *StateDB) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	_ = sender
	_ = dest
	_ = precompiles
	_ = txAccesses
}

func (s *StateDB) AddressInAccessList(addr common.Address) bool {
	_ = addr
	return true
}

func (s *StateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk, slotOk bool) {
	_ = addr
	_ = slot
	return true, true
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (s *StateDB) AddAddressToAccessList(addr common.Address) {
	_ = addr
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (s *StateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	_ = addr
	_ = slot
}

func (s *StateDB) Snapshot() int {
	i := s.ctx.TakeSnapshot()
	s.snapshots[i] = slices.Clone(s.logs)
	return i
}

func (s *StateDB) RevertToSnapshot(i int) {
	s.ctx.RevertToSnapshot(i)
	s.logs = s.snapshots[i]
}

func (s *StateDB) AddLog(log *types.Log) {
	s.logs = append(s.logs, log)
}

func (s *StateDB) GetLogs() []*types.Log {
	return s.logs
}

func (s *StateDB) AddPreimage(common.Hash, []byte) { panic("not implemented") }

// GetTransientState implements vm.StateDB
func (s *StateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return s.transientStorage.Get(addr, key)
}

// SetTransientState implements vm.StateDB
func (s *StateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.transientStorage.Set(addr, key, value)
}

// Prepare implements vm.StateDB
// cleans up refunds, transient storage and "newContract" flags
func (s *StateDB) Prepare(rules params.Rules, sender common.Address, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	// reset refund
	s.refund = 0
	// reset transient storage
	s.transientStorage = newTransientStorage()
	// reset "newContract" flags
	s.newContracts = make(map[common.Address]bool)
	// reset original storage (pre-transaction values)
	s.originalStorage = make(map[common.Address]map[common.Hash]common.Hash)
}

func (s *StateDB) AccessEvents() *state.AccessEvents {
	panic("should not be called")
}

func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	panic("should not be called")
	// TODO: maybe we should "burn" any assets sent to self-destructed accounts
	// here
}

func (s *StateDB) Witness() *stateless.Witness {
	panic("should not be called")
}
