// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

const (
	keyGasRatio     = "g"
	keyEVMOwner     = "o"
	keyNextEVMOwner = "n"
	keyBlockTime    = "b"

	// keyEVMState is the subrealm prefix for the EVM state, used by the emulator
	keyEVMState = "s"
)

func evmStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func setGasRatio(ctx iscp.Sandbox) dict.Dict {
	requireOwner(ctx)
	ctx.State().Set(keyGasRatio, codec.MustDecodeRatio32(ctx.Params().MustGet(evm.FieldGasRatio)).Bytes())
	return nil
}

func getGasRatio(ctx iscp.SandboxView) dict.Dict {
	return result(ctx.State().MustGet(keyGasRatio))
}
