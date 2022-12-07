package transaction

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

func NewRotateChainStateControllerTx(
	aliasID iotago.AliasID,
	newStateController iotago.Address,
	chainOutputID iotago.OutputID,
	chainOutput iotago.Output,
	kp *cryptolib.KeyPair,
) (*iotago.Transaction, error) {
	o, ok := chainOutput.(*iotago.AliasOutput)
	if !ok {
		return nil, fmt.Errorf("provided output is not the correct one. Expected AliasOutput, received %T=%v", chainOutput, chainOutput)
	}
	resolvedAliasID := util.AliasIDFromAliasOutput(o, chainOutputID)
	if resolvedAliasID != aliasID {
		return nil, fmt.Errorf("provided output is not the correct one. Expected ChainID: %s, got: %s",
			aliasID.ToAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
			chainOutput.(*iotago.AliasOutput).AliasID.ToAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
		)
	}

	// create a TX with that UTXO as input, and the updated addr unlock condition on the new output
	inputIDs := iotago.OutputIDs{chainOutputID}
	outSet := iotago.OutputSet{}
	outSet[chainOutputID] = chainOutput
	inputsCommitment := inputIDs.OrderedSet(outSet).MustCommitment()

	newChainOutput := chainOutput.Clone().(*iotago.AliasOutput)
	newChainOutput.AliasID = resolvedAliasID
	oldUnlockConditions := newChainOutput.UnlockConditionSet()
	newChainOutput.Conditions = make(iotago.UnlockConditions, len(oldUnlockConditions))

	// update the unlock conditions to the new state controller
	i := 0
	for t, condition := range oldUnlockConditions {
		newChainOutput.Conditions[i] = condition.Clone()
		if t == iotago.UnlockConditionStateControllerAddress {
			// found the condition to alter
			c, ok := newChainOutput.Conditions[i].(*iotago.StateControllerAddressUnlockCondition)
			if !ok {
				return nil, fmt.Errorf("Unexpected error trying to get StateControllerAddressUnlockCondition")
			}
			c.Address = newStateController
			newChainOutput.Conditions[i] = c.Clone()
		}
		i++
	}

	// remove any "sender feature"
	var newFeatures iotago.Features
	for t, feature := range chainOutput.FeatureSet() {
		if t != iotago.FeatureSender {
			newFeatures = append(newFeatures, feature)
		}
	}
	newChainOutput.Features = newFeatures

	outputs := iotago.Outputs{newChainOutput}
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, kp, parameters.L1().Protocol.NetworkID())
}
