// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
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
	return state.Has(keyPrivileged(addr))
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
	return isc.MustAssetsFromBytes(state.Get(key))
}

func setAllowanceBaseTokens(ctx isc.Sandbox, from, to common.Address, numTokens *big.Int) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		if !numTokens.IsUint64() {
			// Calling `approve(MAX_UINT256)` is semantically equivalent to an "infinite" allowance
			if numTokens.Cmp(gethmath.MaxBig256) == 0 {
				numTokens = big.NewInt(0).SetUint64(math.MaxUint64)
			} else {
				panic("base tokens amount must be an uint64")
			}
		}
		allowance.BaseTokens = numTokens.Uint64()
	})
}

func setAllowanceNativeTokens(ctx isc.Sandbox, from, to common.Address, nativeTokenID iscmagic.NativeTokenID, numTokens *big.Int) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		ntSet := allowance.NativeTokens.MustSet()
		ntSet[nativeTokenID.MustUnwrap()] = &iotago.NativeToken{
			ID:     nativeTokenID.MustUnwrap(),
			Amount: numTokens,
		}
		allowance.NativeTokens = lo.Values(ntSet)
	})
}

func addToAllowance(ctx isc.Sandbox, from, to common.Address, add *isc.Assets) {
	withAllowance(ctx, from, to, func(allowance *isc.Assets) {
		allowance.Add(add)
	})
}

func withAllowance(ctx isc.Sandbox, from, to common.Address, f func(*isc.Assets)) {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	allowance := isc.MustAssetsFromBytes(state.Get(key))
	f(allowance)
	state.Set(key, allowance.Bytes())
}

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Assets) *isc.Assets {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)

	remaining := isc.MustAssetsFromBytes(state.Get(key))
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
	b := state.Get(keyERC20ExternalNativeTokensAddress(nativeTokenID))
	if b == nil {
		return ret, false
	}
	copy(ret[:], b)
	return ret, true
}
