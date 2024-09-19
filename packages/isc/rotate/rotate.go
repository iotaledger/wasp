package rotate

import (
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// IsRotateStateControllerRequest determines if request may be a committee rotation request
func IsRotateStateControllerRequest(req isc.Calldata) bool {
	target := req.Message().Target
	return target.Contract == coreutil.CoreContractGovernanceHname && target.EntryPoint == coreutil.CoreEPRotateStateControllerHname
}

func NewRotateRequestOffLedger(chainID isc.ChainID, newStateAddress *cryptolib.Address, keyPair *cryptolib.KeyPair, gasBudget uint64) isc.Request {
	args := isc.NewCallArguments(codec.Address.Encode(newStateAddress))
	nonce := uint64(time.Now().UnixNano())

	ret := isc.NewOffLedgerRequest(chainID, isc.NewMessage(coreutil.CoreContractGovernanceHname, coreutil.CoreEPRotateStateControllerHname, args), nonce, gasBudget)
	return ret.Sign(keyPair)
}
