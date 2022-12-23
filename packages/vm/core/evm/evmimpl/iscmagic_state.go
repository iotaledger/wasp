// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

const (
	prefixPrivileged = "p"
	prefixAllowance  = "a"
)

// directory of EVM contracts that have access to the privileged methods of ISC magic
func keyPrivileged(addr common.Address) kv.Key {
	return kv.Key(prefixPrivileged) + kv.Key(addr.Bytes())
}

func isCallerPrivileged(ctx isc.SandboxBase, addr common.Address) bool {
	state := iscMagicSubrealmR(ctx.StateR())
	return state.MustHas(keyPrivileged(addr))
}

func addToPrivileged(ctx isc.Sandbox, addr common.Address) {
	state := iscMagicSubrealm(ctx.State())
	state.Set(keyPrivileged(addr), []byte{1})
}

// allowance between two EVM accounts
func keyAllowance(from, to common.Address) kv.Key {
	return kv.Key(prefixAllowance) + kv.Key(from.Bytes()) + kv.Key(to.Bytes())
}

func getAllowance(ctx isc.SandboxBase, from, to common.Address) *isc.Allowance {
	state := iscMagicSubrealmR(ctx.StateR())
	key := keyAllowance(from, to)
	return codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
}

func addToAllowance(ctx isc.Sandbox, from, to common.Address, add *isc.Allowance) {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	allowance := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
	allowance.Add(add)
	state.Set(key, allowance.Bytes())
}

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Allowance) *isc.Allowance {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)

	remaining := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
	if taken.IsEmpty() {
		taken = remaining.Clone()
	}

	ok := remaining.SpendFromBudget(taken)
	ctx.Requiref(ok, "takeAllowedFunds: not previously allowed")
	if remaining.IsEmpty() {
		state.Del(key)
	} else {
		state.Set(key, remaining.Bytes())
	}

	return taken
}
