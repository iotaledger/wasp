package transaction

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
)

func NewRotateChainStateControllerTx(
	chainID iotago.AliasID,
	newStateController iotago.Address,
	chainOutputID iotago.OutputID,
	chainOutput iotago.Output,
	kp *cryptolib.KeyPair,
) (*iotago.Transaction, error) {
	if o, ok := chainOutput.(*iotago.AliasOutput); !ok || o.AliasID != chainID {
		return nil, fmt.Errorf("provided output is not the correct one. expected ChainID: %s", chainID.ToAddress().Bech32(parameters.L1().Protocol.Bech32HRP))
	}

	// create a TX with that UTXO as input, and the updated addr unlock condition on the new output
	inputIDs := iotago.OutputIDs{chainOutputID}
	outSet := iotago.OutputSet{}
	outSet[chainOutputID] = chainOutput
	inputsCommitment := inputIDs.OrderedSet(outSet).MustCommitment()

	newChainOutput := chainOutput.Clone().(*iotago.AliasOutput)
	oldUnlockConditions := newChainOutput.UnlockConditionSet()
	newChainOutput.Conditions = make(iotago.UnlockConditions, len(oldUnlockConditions))
	i := 0
	for t, condition := range oldUnlockConditions {
		newChainOutput.Conditions[i] = condition.Clone()
		if t != iotago.UnlockConditionStateControllerAddress {
			i++
			continue
		}
		// found the condition to alter
		c, ok := newChainOutput.Conditions[i].(*iotago.StateControllerAddressUnlockCondition)
		if !ok {
			return nil, fmt.Errorf("Unexpected error trying to get StateControllerAddressUnlockCondition")
		}
		c.Address = newStateController
		newChainOutput.Conditions[i] = c.Clone()
		i++
	}

	outputs := iotago.Outputs{newChainOutput}
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, kp, parameters.L1().Protocol.NetworkID())
}
