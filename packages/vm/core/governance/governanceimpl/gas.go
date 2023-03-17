// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// setFeePolicy sets the global fee policy for the chain in serialized form
// Input:
// - governance.ParamFeePolicyBytes must contain bytes of the policy record
func setFeePolicy(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	data := ctx.Params().MustGetBytes(governance.ParamFeePolicyBytes)
	_, err := gas.FeePolicyFromBytes(data)
	ctx.RequireNoError(err)

	ctx.State().Set(governance.VarGasFeePolicyBytes, data)
	return nil
}

// getFeeInfo returns fee policy in serialized form
func getFeePolicy(ctx isc.SandboxView) dict.Dict {
	gp := governance.MustGetGasFeePolicy(ctx.StateR())

	ret := dict.New()
	ret.Set(governance.ParamFeePolicyBytes, gp.Bytes())
	return ret
}

func setEVMGasRatio(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	ratio := codec.MustDecodeRatio32(ctx.Params().MustGet(governance.ParamEVMGasRatio))
	ctx.Requiref(!ratio.HasZeroComponent(), "invalid ratio")
	policy := governance.MustGetGasFeePolicy(ctx.StateR())
	policy.EVMGasRatio = ratio
	ctx.State().Set(governance.VarGasFeePolicyBytes, policy.Bytes())
	return nil
}

func getEVMGasRatio(ctx isc.SandboxView) dict.Dict {
	policy := governance.MustGetGasFeePolicy(ctx.StateR())
	return dict.Dict{
		governance.ParamEVMGasRatio: policy.EVMGasRatio.Bytes(),
	}
}

func setGasLimits(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	data := ctx.Params().MustGetBytes(governance.ParamGasLimitsBytes)
	_, err := gas.LimitsFromBytes(data)
	ctx.RequireNoError(err)

	ctx.State().Set(governance.VarGasLimitsBytes, data)
	return nil
}

func getGasLimits(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		governance.ParamGasLimitsBytes: governance.MustGetGasLimits(ctx.StateR()).Bytes(),
	}
}
