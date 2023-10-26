// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	keyAccountNonce          = "n"
	keyAccountCode           = "c"
	keyAccountState          = "s"
	keyAccountSelfDestructed = "S"
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
	ctx       Context
	kv        kv.KVStore // subrealm of ctx.State()
	logs      []*types.Log
	snapshots map[int][]*types.Log
	refund    uint64
}

var _ vm.StateDB = &StateDB{}

func NewStateDB(ctx Context) *StateDB {
	return &StateDB{
		ctx:       ctx,
		kv:        StateDBSubrealm(ctx.State()),
		snapshots: make(map[int][]*types.Log),
	}
}

func CreateAccount(kv kv.KVStore, addr common.Address) {
	SetNonce(kv, addr, 0)
}

func (s *StateDB) CreateAccount(addr common.Address) {
	CreateAccount(s.kv, addr)
}

func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	if amount.Sign() == -1 {
		panic("unexpected negative amount")
	}
	s.ctx.SubBaseTokensBalance(addr, util.EthereumDecimalsToBaseTokenDecimals(amount, s.ctx.BaseTokensDecimals()))
}

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	if amount.Sign() == -1 {
		panic("unexpected negative amount")
	}
	s.ctx.AddBaseTokensBalance(addr, util.EthereumDecimalsToBaseTokenDecimals(amount, s.ctx.BaseTokensDecimals()))
}

func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	return util.BaseTokensDecimalsToEthereumDecimals(s.ctx.GetBaseTokensBalance(addr), s.ctx.BaseTokensDecimals())
}

func GetNonce(s kv.KVStoreReader, addr common.Address) uint64 {
	return codec.MustDecodeUint64(s.Get(accountNonceKey(addr)), 0)
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
	kv.Set(accountNonceKey(addr), codec.EncodeUint64(n))
}

func (s *StateDB) SetNonce(addr common.Address, n uint64) {
	SetNonce(s.kv, addr, n)
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	// TODO cache the code hash?
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

func (s *StateDB) SetCode(addr common.Address, code []byte) {
	SetCode(s.kv, addr, code)
}

func (s *StateDB) GetCodeSize(addr common.Address) int {
	// TODO cache the code size?
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

func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
	SetState(s.kv, addr, key, value)
}

func (s *StateDB) SelfDestruct(addr common.Address) {
	if !s.Exist(addr) {
		return
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
	s.ctx.SubBaseTokensBalance(addr, s.ctx.GetBaseTokensBalance(addr))

	s.kv.Set(accountSelfDestructedKey(addr), []byte{1})
}

func (s *StateDB) HasSelfDestructed(addr common.Address) bool {
	return s.kv.Has(accountSelfDestructedKey(addr))
}

func (s *StateDB) Selfdestruct6780(addr common.Address) {
	panic("unimplemented")
}

// Exist reports whether the given account exists in state.
// Notably this should also return true for self-destructed accounts.
func (s *StateDB) Exist(addr common.Address) bool {
	return s.kv.Has(accountNonceKey(addr))
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

func (s *StateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) error {
	panic("not implemented")
}

// GetTransientState implements vm.StateDB
func (*StateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	panic("unimplemented")
}

// Prepare implements vm.StateDB
func (s *StateDB) Prepare(rules params.Rules, sender common.Address, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	// do nothing
}

// SetTransientState implements vm.StateDB
func (*StateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	panic("unimplemented")
}
