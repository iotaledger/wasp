// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// setFeePolicy sets the global fee policy for the chain in serialized form
// Input:
// - governance.ParamFeePolicyBytes must contain bytes of the policy record
func setFeePolicy(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner("governance.setFeePolicy: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	data := params.MustGetBytes(governance.ParamFeePolicyBytes)
	_, err := gas.GasFeePolicyFromBytes(data)
	ctx.RequireNoError(err)

	ctx.State().Set(governance.VarGasFeePolicyBytes, data)
	return nil
}

// getFeeInfo returns fee policy in serialized form
func getFeePolicy(ctx iscp.SandboxView) dict.Dict {
	gp := governance.GetGasFeePolicy(ctx.State())

	ret := dict.New()
	ret.Set(governance.ParamFeePolicyBytes, gp.Bytes())
	return ret
}
