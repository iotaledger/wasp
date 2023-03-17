package rotate

import (
	"time"

	"github.com/iotaledger/hive.go/crypto/identity"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
)

// IsRotateStateControllerRequest determines if request may be a committee rotation request
func IsRotateStateControllerRequest(req isc.Calldata) bool {
	target := req.CallTarget()
	return target.Contract == coreutil.CoreContractGovernanceHname && target.EntryPoint == coreutil.CoreEPRotateStateControllerHname
}

func NewRotateRequestOffLedger(chainID isc.ChainID, newStateAddress iotago.Address, keyPair *cryptolib.KeyPair, gasBudget uint64) isc.Request {
	args := dict.New()
	args.Set(coreutil.ParamStateControllerAddress, codec.EncodeAddress(newStateAddress))
	nonce := uint64(time.Now().UnixNano())
	ret := isc.NewOffLedgerRequest(chainID, coreutil.CoreContractGovernanceHname, coreutil.CoreEPRotateStateControllerHname, args, nonce, gasBudget)
	return ret.Sign(keyPair)
}

func MakeRotateStateControllerTransaction(
	nextAddr iotago.Address,
	chainInput *isc.AliasOutputWithID,
	ts time.Time,
	accessPledge, consensusPledge identity.ID,
) (*iotago.TransactionEssence, error) {
	output := chainInput.GetAliasOutput().Clone().(*iotago.AliasOutput)
	for i := range output.Conditions {
		if _, ok := output.Conditions[i].(*iotago.StateControllerAddressUnlockCondition); ok {
			output.Conditions[i] = &iotago.StateControllerAddressUnlockCondition{Address: nextAddr}
		}
		// TODO: it is probably not the correct way to do the governance transition
		if _, ok := output.Conditions[i].(*iotago.GovernorAddressUnlockCondition); ok {
			output.Conditions[i] = &iotago.GovernorAddressUnlockCondition{Address: nextAddr}
		}
	}

	// remove any "sender feature"
	var newFeatures iotago.Features
	for t, feature := range chainInput.GetAliasOutput().FeatureSet() {
		if t != iotago.FeatureSender {
			newFeatures = append(newFeatures, feature)
		}
	}
	output.Features = newFeatures

	result := &iotago.TransactionEssence{
		NetworkID: parameters.L1().Protocol.NetworkID(),
		Inputs:    iotago.Inputs{chainInput.OutputID().UTXOInput()},
		Outputs:   iotago.Outputs{output},
		Payload:   nil,
	}
	inputsCommitment := iotago.Outputs{chainInput.GetAliasOutput()}.MustCommitment()
	copy(result.InputsCommitment[:], inputsCommitment)
	return result, nil
}
