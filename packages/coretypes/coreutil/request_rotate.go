package coreutil

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
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

func MakeRotateStateControllerTransaction(
	nextAddr ledgerstate.Address,
	chainInput *ledgerstate.AliasOutput,
	req ledgerstate.Output,
	ts time.Time,
	accessPledge, consensusPledge identity.ID,
) (*ledgerstate.TransactionEssence, error) {

	inputs := []ledgerstate.Output{chainInput}
	if req != nil {
		inputs = append(inputs, req)
	}
	txb := utxoutil.NewBuilder(inputs...).
		WithTimestamp(ts).
		WithAccessPledge(accessPledge).
		WithConsensusPledge(consensusPledge)
	chained := chainInput.NewAliasOutputNext(true)
	if err := chained.SetStateAddress(nextAddr); err != nil {
		return nil, err
	}
	if err := txb.ConsumeAliasInput(chainInput.Address()); err != nil {
		return nil, err
	}
	if err := txb.AddOutputAndSpendUnspent(chained); err != nil {
		return nil, err
	}
	essence, _, err := txb.BuildEssence()
	if err != nil {
		return nil, err
	}
	return essence, nil
}
