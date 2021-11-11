package rotate

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// IsRotateStateControllerRequest determines if request may be a committee rotation request
func IsRotateStateControllerRequest(req iscp.Request) bool {
	target := req.Target()
	return target.Contract == coreutil.CoreContractGovernanceHname && target.EntryPoint == coreutil.CoreEPRotateStateControllerHname
}

func NewRotateRequestOffLedger(chainID *iscp.ChainID, newStateAddress ledgerstate.Address, keyPair *ed25519.KeyPair) *request.OffLedger {
	args := requestargs.New(nil)
	args.AddEncodeSimple(coreutil.ParamStateControllerAddress, codec.EncodeAddress(newStateAddress))
	ret := request.NewOffLedger(chainID, coreutil.CoreContractGovernanceHname, coreutil.CoreEPRotateStateControllerHname, args)
	ret.Sign(keyPair)
	return ret
}

func MakeRotateStateControllerTransaction(
	nextAddr ledgerstate.Address,
	chainInput *ledgerstate.AliasOutput,
	ts time.Time,
	accessPledge, consensusPledge identity.ID,
) (*ledgerstate.TransactionEssence, error) {
	txb := utxoutil.NewBuilder(chainInput).
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
