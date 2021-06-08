package coreutil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// IsRotateCommitteeRequest determines if request may be a committee rotation request
func IsRotateCommitteeRequest(req coretypes.Request) bool {
	targetContract, targetEP := req.Target()
	if targetContract != CoreContractGovernanceHname {
		return false
	}
	if targetEP != CoreEPRotateCommitteeHname {
		return false
	}
	par, ok := req.Params()
	if !ok {
		return false
	}
	deco := kvdecoder.New(par)
	if a, err := deco.GetAddress(ParamStateAddress); a == nil || err != nil {
		return false
	}
	return true
}
