package coreutil

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// IsRotateCommitteeRequest determines if request may be a committee rotation request
func IsRotateCommitteeRequest(req coretypes.Request) bool {
	targetContract, targetEP := req.Target()
	return targetContract == CoreContractGovernanceHname && targetEP == CoreEPRotateCommitteeHname
}

func NewRotateRequestOffLedger(newStateAddress ledgerstate.Address, keyPair *ed25519.KeyPair) coretypes.Request {
	args := requestargs.New(nil)
	args.AddEncodeSimple(ParamStateControllerAddress, codec.EncodeAddress(newStateAddress))
	ret := request.NewRequestOffLedger(CoreContractGovernanceHname, CoreEPRotateCommitteeHname, args)
	ret.Sign(keyPair)
	return ret
}
