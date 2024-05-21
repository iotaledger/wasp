// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// setFeePolicy sets the global fee policy for the chain in serialized form
func setFeePolicy(ctx isc.Sandbox, fp *gas.FeePolicy) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetGasFeePolicy(fp)
	return nil
}

func getFeePolicy(ctx isc.SandboxView) *gas.FeePolicy {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetGasFeePolicy()
}

var errInvalidGasRatio = coreerrors.Register("invalid gas ratio").Create()

func setEVMGasRatio(ctx isc.Sandbox, ratio util.Ratio32) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	if !ratio.IsValid() {
		panic(errInvalidGasRatio)
	}
	state := governance.NewStateWriterFromSandbox(ctx)
	policy := state.GetGasFeePolicy()
	policy.EVMGasRatio = ratio
	state.SetGasFeePolicy(policy)
	return nil
}

func getEVMGasRatio(ctx isc.SandboxView) util.Ratio32 {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetGasFeePolicy().EVMGasRatio
}

func setGasLimits(ctx isc.Sandbox, limits *gas.Limits) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetGasLimits(limits)
	return nil
}

func getGasLimits(ctx isc.SandboxView) *gas.Limits {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetGasLimits()
}
