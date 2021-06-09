package coreutil

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// IsSolidRotateCommitteeRequest determines if request may be a committee rotation request
func IsSolidRotateCommitteeRequest(req coretypes.Request) bool {
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

func NewRotateRequestOffLedger(newStateAddress ledgerstate.Address, keyPair *ed25519.KeyPair) coretypes.Request {
	args := requestargs.New(nil)
	args.AddEncodeSimple(ParamStateAddress, codec.EncodeAddress(newStateAddress))
	ret := request.NewRequestOffLedger(CoreContractGovernanceHname, CoreEPRotateCommitteeHname, args)
	ret.Sign(keyPair)
	return ret
}
