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
	unspentOutputs iotago.OutputSet,
	kp *cryptolib.KeyPair,
) (*iotago.Transaction, error) {
	// search for the UTXO that has the CHAIN ID

	for id, utxo := range unspentOutputs {
		if utxo.UnlockConditionSet().ImmutableAlias().Address.AliasID() != chainID {
			continue
		}
		// found the desired output

		// create a TX with that UTXO as input, and the updated addr unlock condition on the new UTXO
		inputIDs := iotago.OutputIDs{id}
		inputsCommitment := inputIDs.OrderedSet(unspentOutputs).MustCommitment()

		newChainOutput := utxo.Clone().(*iotago.AliasOutput)
		oldUnlockConditions := utxo.UnlockConditionSet()
		newUnlockConditions := make(iotago.UnlockConditions, len(oldUnlockConditions))
		for i, condition := range oldUnlockConditions {
			if condition.Type() == iotago.UnlockConditionStateControllerAddress {
				condition.(*iotago.AddressUnlockCondition) // TODO
			}
			newUnlockConditions[i] = oldUnlockConditions[i]
		}

		newChainOutput.Conditions = newUnlockConditions
		newChainOutput.UnlockConditionSet().StateControllerAddress()
		outputs := iotago.Outputs{newChainOutput}

		return CreateAndSignTx(inputIDs, inputsCommitment, outputs, kp, parameters.L1.Protocol.NetworkID())
	}

	return nil, fmt.Errorf("UTXO with chainID %s not found", chainID.ToAddress().Bech32(parameters.L1.Protocol.Bech32HRP))
}
