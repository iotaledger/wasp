// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/samber/lo"

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
	ctx.State().Set(governance.VarGasFeePolicyBytes, fp.Bytes())
	return nil
}

func getFeePolicy(ctx isc.SandboxView) *gas.FeePolicy {
	return lo.Must(governance.GetGasFeePolicy(ctx.StateR()))
}

var errInvalidGasRatio = coreerrors.Register("invalid gas ratio").Create()

func setEVMGasRatio(ctx isc.Sandbox, ratio util.Ratio32) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	if !ratio.IsValid() {
		panic(errInvalidGasRatio)
	}
	policy := governance.MustGetGasFeePolicy(ctx.StateR())
	policy.EVMGasRatio = ratio
	ctx.State().Set(governance.VarGasFeePolicyBytes, policy.Bytes())
	return nil
}

func getEVMGasRatio(ctx isc.SandboxView) util.Ratio32 {
	return lo.Must(governance.GetGasFeePolicy(ctx.StateR())).EVMGasRatio
}

func setGasLimits(ctx isc.Sandbox, limits *gas.Limits) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	ctx.State().Set(governance.VarGasLimitsBytes, limits.Bytes())
	return nil
}

func getGasLimits(ctx isc.SandboxView) *gas.Limits {
	return lo.Must(governance.GetGasLimits(ctx.StateR()))
}
