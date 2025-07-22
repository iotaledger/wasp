// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
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
	b := state.Get(key)
	if b == nil {
		return isc.NewEmptyAssets()
	}
	return lo.Must(isc.AssetsFromBytes(b))
}

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
	allowance := getAllowance(ctx, from, to)
	f(allowance)
	state := evm.ISCMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	state.Set(key, allowance.Bytes())
}

var errFundsNotAllowed = coreerrors.Register("remaining allowance insufficient").Create()

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Assets) {
	remaining := getAllowance(ctx, from, to)
	if ok := remaining.Spend(taken); !ok {
		panic(errFundsNotAllowed)
	}
	state := evm.ISCMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	if remaining.IsEmpty() {
		state.Del(key)
	} else {
		state.Set(key, remaining.Bytes())
	}
}
