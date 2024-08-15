// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

// The ISC magic contract stores some data in the ISC state.
const (
	// prefixPrivileged stores the directory of EVM contracts that have access to
	// the "privileged" ISC magic methods.
	// Covered in: TestStorageContract
	prefixPrivileged = "p"
	// prefixAllowance stores the allowance between accounts (e.g. by calling
	// ISC.allow() from solidity).
	// Covered in: TestSendBaseTokens
	prefixAllowance = "a"
	// prefixERC20ExternalNativeTokens stores the directory of ERC20 contracts
	// registered by calling ISC.registerERC20NativeToken() from solidity.
	// Covered in: TestERC20CoinWithExternalFoundry
	prefixERC20ExternalNativeTokens = "e"
)

// directory of EVM contracts that have access to the privileged methods of ISC magic
func keyPrivileged(addr common.Address) kv.Key {
	return prefixPrivileged + kv.Key(addr.Bytes())
}

func isCallerPrivileged(ctx isc.SandboxBase, addr common.Address) bool {
	state := evm.ISCMagicSubrealmR(ctx.StateR())
	return state.Has(keyPrivileged(addr))
}

func addToPrivileged(evmState kv.KVStore, addr common.Address) {
	state := evm.ISCMagicSubrealm(evmState)
	state.Set(keyPrivileged(addr), []byte{1})
}

// allowance between two EVM accounts
func keyAllowance(from, to common.Address) kv.Key {
	return prefixAllowance + kv.Key(from.Bytes()) + kv.Key(to.Bytes())
}

func getAllowance(ctx isc.SandboxBase, from, to common.Address) *isc.Assets {
	state := evm.ISCMagicSubrealmR(ctx.StateR())
	key := keyAllowance(from, to)
	return lo.Must(isc.AssetsFromBytes(state.Get(key)))
}

var errBaseTokensMustBeUint64 = coreerrors.Register("base tokens amount must be an uint64").Create()

func setAllowanceBaseTokens(ctx isc.Sandbox, from, to common.Address, amount coin.Value) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		allowance.SetBaseTokens(amount)
	})
}

func setAllowanceCoin(ctx isc.Sandbox, from, to common.Address, coinType coin.Type, amount coin.Value) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		allowance.Coins.Set(coinType, amount)
	})
}

func addToAllowance(ctx isc.Sandbox, from, to common.Address, add *isc.Assets) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		allowance.Add(add)
	})
}

func withAllowance(ctx isc.Sandbox, from, to common.Address, f func(*isc.Assets)) {
	state := evm.ISCMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	allowance := lo.Must(isc.AssetsFromBytes(state.Get(key)))
	f(allowance)
	state.Set(key, allowance.Bytes())
}

var errFundsNotAllowed = coreerrors.Register("remaining allowance insufficient").Create()

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Assets) {
	state := evm.ISCMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	remaining := lo.Must(isc.AssetsFromBytes(state.Get(key)))
	if ok := remaining.Spend(taken); !ok {
		panic(errFundsNotAllowed)
	}
	if remaining.IsEmpty() {
		state.Del(key)
	} else {
		state.Set(key, remaining.Bytes())
	}
}

// directory of ERC20 contract addresses by native token ID
func keyERC20ExternalNativeTokensAddress(nativeTokenID coin.Type) kv.Key {
	return prefixERC20ExternalNativeTokens + kv.Key(nativeTokenID[:])
}

func addERC20ExternalNativeTokensAddress(ctx isc.Sandbox, nativeTokenID coin.Type, addr common.Address) {
	state := evm.ISCMagicSubrealm(ctx.State())
	state.Set(keyERC20ExternalNativeTokensAddress(nativeTokenID), addr.Bytes())
}

func getERC20ExternalNativeTokensAddress(ctx isc.SandboxBase, nativeTokenID coin.Type) (ret common.Address, ok bool) {
	state := evm.ISCMagicSubrealmR(ctx.StateR())
	b := state.Get(keyERC20ExternalNativeTokensAddress(nativeTokenID))
	if b == nil {
		return ret, false
	}
	copy(ret[:], b)
	return ret, true
}

// findERC20NativeTokenContractAddress returns the address of an
// ERC20Coin or ERC20ExternalNativeTokens contract.
func findERC20NativeTokenContractAddress(ctx isc.Sandbox, coinType coin.Type) (common.Address, bool) {
	panic("refactor me: findERC20NativeTokenContractAddress evm (is this still even needed?)")
	/*
		addr, ok := getERC20ExternalNativeTokensAddress(ctx, coinType)
		if ok {
			return addr, true
		}
		addr = iscmagic.ERC20CoinAddress(coinType.FoundrySerialNumber())
		stateDB := emulator.NewStateDB(newEmulatorContext(ctx))
		if stateDB.Exist(addr) {
			return addr, true
		}
		return common.Address{}, false
	*/
}
