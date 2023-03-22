// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
)

const (
	prefixPrivileged                = "p"
	prefixAllowance                 = "a"
	prefixERC20ExternalNativeTokens = "e"
)

// directory of EVM contracts that have access to the privileged methods of ISC magic
func keyPrivileged(addr common.Address) kv.Key {
	return kv.Key(prefixPrivileged) + kv.Key(addr.Bytes())
}

func isCallerPrivileged(ctx isc.SandboxBase, addr common.Address) bool {
	state := iscMagicSubrealmR(ctx.StateR())
	return state.MustHas(keyPrivileged(addr))
}

func addToPrivileged(s kv.KVStore, addr common.Address) {
	state := iscMagicSubrealm(s)
	state.Set(keyPrivileged(addr), []byte{1})
}

// allowance between two EVM accounts
func keyAllowance(from, to common.Address) kv.Key {
	return kv.Key(prefixAllowance) + kv.Key(from.Bytes()) + kv.Key(to.Bytes())
}

func getAllowance(ctx isc.SandboxBase, from, to common.Address) *isc.Assets {
	state := iscMagicSubrealmR(ctx.StateR())
	key := keyAllowance(from, to)
	return isc.MustAssetsFromBytes(state.MustGet(key))
}

func addToAllowance(ctx isc.Sandbox, from, to common.Address, add *isc.Assets) {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	allowance := isc.MustAssetsFromBytes(state.MustGet(key))
	allowance.Add(add)
	state.Set(key, allowance.Bytes())
}

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Assets) *isc.Assets {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)

	remaining := isc.MustAssetsFromBytes(state.MustGet(key))
	if taken.IsEmpty() {
		taken = remaining.Clone()
	}

	ok := remaining.Spend(taken)
	ctx.Requiref(ok, "takeAllowedFunds: not previously allowed")
	if remaining.IsEmpty() {
		state.Del(key)
	} else {
		state.Set(key, remaining.Bytes())
	}

	return taken
}

// directory of ERC20 contract addresses by native token ID
func keyERC20ExternalNativeTokensAddress(nativeTokenID iotago.NativeTokenID) kv.Key {
	return kv.Key(prefixERC20ExternalNativeTokens) + kv.Key(nativeTokenID[:])
}

func addERC20ExternalNativeTokensAddress(ctx isc.Sandbox, nativeTokenID iotago.NativeTokenID, addr common.Address) {
	state := iscMagicSubrealm(ctx.State())
	state.Set(keyERC20ExternalNativeTokensAddress(nativeTokenID), addr.Bytes())
}

func getERC20ExternalNativeTokensAddress(ctx isc.SandboxBase, nativeTokenID iotago.NativeTokenID) (ret common.Address, ok bool) {
	state := iscMagicSubrealmR(ctx.StateR())
	b := state.MustGet(keyERC20ExternalNativeTokensAddress(nativeTokenID))
	if b == nil {
		return ret, false
	}
	copy(ret[:], b)
	return ret, true
}
