// Package transaction provides functionality for creating and managing transactions in the Wasp ecosystem.
package transaction

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func NewChangeGovControllerTx(
	chainID isc.ChainID,
	newGovController iotago.Address,
	// utxos iotago.OutputSet,
	wallet cryptolib.Signer,
) (*iotago.ObjectID, error) {
	panic("refactor me: Reimplement GovController")
	/*
		// find the correct chain UTXO
		var chainOutput *iotago.AliasOutput
		var chainOutputID iotago.OutputID
		for id, o := range utxos {
			ao, ok := o.(*iotago.AliasOutput)
			if !ok {
				continue
			}
			if util.AliasIDFromAliasOutput(ao, id) == chainID {
				chainOutputID = id
				chainOutput = ao.Clone().(*iotago.AliasOutput)
				break
			}
		}
		if chainOutput == nil {
			return nil, fmt.Errorf("unable to find UTXO for chain (%s) in owned UTXOs", chainID.String())
		}

		newConditions := make(iotago.UnlockConditions, len(chainOutput.Conditions))
		for i, c := range chainOutput.Conditions {
			if _, ok := c.(*iotago.GovernorAddressUnlockCondition); ok {
				// change the gov unlock condition to the new owner
				newConditions[i] = &iotago.GovernorAddressUnlockCondition{
					Address: newGovController,
				}
				continue
			}
			newConditions[i] = c
		}
		chainOutput.Conditions = newConditions
		chainOutput.AliasID = chainID // in case right after mint where outputID is still 0


		inputIDs := iotago.OutputIDs{chainOutputID}

		inputsCommitment := inputIDs.OrderedSet(utxos).MustCommitment()
		outputs := []iotago.Output{chainOutput}
	*/
	// return nil, errors.New("iotago.Outputs are not supported anymore") // CreateAndSignTx(inputIDs.UTXOInputs(), inputsCommitment, outputs, wallet, parameters.L1().Protocol.NetworkID())
}
