// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// setContractFee sets fee for the particular smart contract
// Input:
// - ParamHname iscp.Hname smart contract ID
// - ParamOwnerFee int64 non-negative value of the owner fee. May be skipped, then it is not set
// - ParamValidatorFee int64 non-negative value of the contract fee. May be skipped, then it is not set
func setContractFee(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(governance.CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "governance.setContractFee: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	hname := params.MustGetHname(governance.ParamHname)
	rec := governance.FindContractFees(ctx.State(), hname)
	if rec == nil {
		rec = governance.NewContractFeesRecord(0, 0)
	}

	ownerFee := params.MustGetInt64(governance.ParamOwnerFee, -1)
	ownerFeeSet := ownerFee >= 0
	validatorFee := params.MustGetInt64(governance.ParamValidatorFee, -1)
	validatorFeeSet := validatorFee >= 0

	a.Require(ownerFeeSet || validatorFeeSet, "governance.setContractFee: wrong parameters")
	if ownerFeeSet {
		rec.OwnerFee = uint64(ownerFee)
	}
	if validatorFeeSet {
		rec.ValidatorFee = uint64(validatorFee)
	}
	collections.NewMap(ctx.State(), governance.VarContractFeesRegistry).MustSetAt(hname.Bytes(), rec.Bytes())
	return nil, nil
}

// getFeeInfo returns fee information for the contract.
// Input:
// - ParamHname iscp.Hname contract id
// Output:
// - ParamFeeColor ledgerstate.Color color of tokens accepted for fees
// - ParamValidatorFee int64 minimum fee for contract
// Note: return default chain values if contract doesn't exist
func getFeeInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	hname, err := params.GetHname(governance.ParamHname)
	if err != nil {
		return nil, err
	}
	feeColor, ownerFee, validatorFee := governance.GetFeeInfo(ctx, hname)
	ret := dict.New()
	ret.Set(governance.ParamFeeColor, codec.EncodeColor(feeColor))
	ret.Set(governance.ParamOwnerFee, codec.EncodeUint64(ownerFee))
	ret.Set(governance.ParamValidatorFee, codec.EncodeUint64(validatorFee))
	return ret, nil
}
