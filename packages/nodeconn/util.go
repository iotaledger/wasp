package nodeconn

import (
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/core/generics/shrinkingmap"
	"github.com/iotaledger/hive.go/serializer/v2"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func getAliasID(outputID iotago.OutputID, aliasOutput *iotago.AliasOutput) iotago.AliasID {
	if aliasOutput.AliasID.Empty() {
		return iotago.AliasIDFromOutputID(outputID)
	}

	return aliasOutput.AliasID
}

func outputInfoFromINXOutput(output *inx.LedgerOutput) (*isc.OutputInfo, error) {
	outputID := output.UnwrapOutputID()

	iotaOutput, err := output.UnwrapOutput(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}

	return isc.NewOutputInfo(outputID, iotaOutput, iotago.TransactionID{}), nil
}

func outputInfoFromINXSpent(spent *inx.LedgerSpent) (*isc.OutputInfo, error) {
	outputInfo, err := outputInfoFromINXOutput(spent.GetOutput())
	if err != nil {
		return nil, err
	}

	outputInfo.TransactionIDSpent = spent.UnwrapTransactionIDSpent()
	return outputInfo, nil
}

func unwrapOutputs(outputs []*inx.LedgerOutput) ([]*isc.OutputInfo, error) {
	result := make([]*isc.OutputInfo, len(outputs))

	for i := range outputs {
		outputInfo, err := outputInfoFromINXOutput(outputs[i])
		if err != nil {
			return nil, err
		}
		result[i] = outputInfo
	}

	return result, nil
}

func unwrapSpents(spents []*inx.LedgerSpent) ([]*isc.OutputInfo, error) {
	result := make([]*isc.OutputInfo, len(spents))

	for i := range spents {
		outputInfo, err := outputInfoFromINXSpent(spents[i])
		if err != nil {
			return nil, err
		}
		result[i] = outputInfo
	}

	return result, nil
}

// wasOutputIDConsumedBefore checks recursively if "targetOutputID" was consumed before "outputID"
// by walking the consumed outputs of the transaction that created "outputID".
func wasOutputIDConsumedBefore(consumedOutputsMapByTransactionID map[iotago.TransactionID]map[iotago.OutputID]struct{}, targetOutputID iotago.OutputID, outputID iotago.OutputID) bool {
	consumedOutputs, exists := consumedOutputsMapByTransactionID[outputID.TransactionID()]
	if !exists {
		// if the transaction that created the "outputID" was not part of the milestone, the "outputID" was consumed before "targetOutputID"
		return false
	}

	for consumedOutput := range consumedOutputs {
		if consumedOutput == targetOutputID {
			// we found the "targetOutputID" in the past of "outputID"
			return true
		}

		// walk all consumed outputs of that transaction recursively
		if wasOutputIDConsumedBefore(consumedOutputsMapByTransactionID, targetOutputID, consumedOutput) {
			return true
		}
	}

	// we didn't find the "targetOutputID" in the past of "outputID"
	return false
}

func sortAliasOutputsOfChain(trackedChainAliasOutputsCreated []*isc.OutputInfo, trackedAliasOutputsConsumedMapByTransactionID map[iotago.TransactionID]map[iotago.OutputID]struct{}) error {
	var innerErr error

	sort.SliceStable(trackedChainAliasOutputsCreated, func(i, j int) bool {
		outputInfo1 := trackedChainAliasOutputsCreated[i]
		outputInfo2 := trackedChainAliasOutputsCreated[j]

		aliasOutput1 := outputInfo1.Output.(*iotago.AliasOutput)
		aliasOutput2 := outputInfo2.Output.(*iotago.AliasOutput)

		// check if state indexes are equal.
		if aliasOutput1.StateIndex != aliasOutput2.StateIndex {
			return aliasOutput1.StateIndex < aliasOutput2.StateIndex
		}

		outputID1 := outputInfo1.OutputID
		outputID2 := outputInfo2.OutputID

		// in case of a governance transition, the state index is equal.
		if !outputInfo1.Consumed() {
			if !outputInfo2.Consumed() {
				// this should never happen because there can't be two alias outputs with the same alias ID that are unspent.
				innerErr = fmt.Errorf("two unspent alias outputs with same AliasID found (Output1: %s, Output2: %s", outputID1.ToHex(), outputID2.ToHex())
			}
			return false
		}

		if !outputInfo2.Consumed() {
			// first output was consumed, second was not, so first is before second.
			return true
		}

		// we need to figure out the order in which they were consumed (recursive).
		if wasOutputIDConsumedBefore(trackedAliasOutputsConsumedMapByTransactionID, outputID1, outputID2) {
			return true
		}

		if wasOutputIDConsumedBefore(trackedAliasOutputsConsumedMapByTransactionID, outputID2, outputID1) {
			return false
		}

		innerErr = fmt.Errorf("two consumed alias outputs with same AliasID found, but ordering is unclear (Output1: %s, Output2: %s", outputID1.ToHex(), outputID2.ToHex())
		return false
	})

	return innerErr
}

func getAliasIDAliasOutput(outputInfo *isc.OutputInfo) iotago.AliasID {
	if outputInfo.Output.Type() != iotago.OutputAlias {
		return iotago.AliasID{}
	}

	return getAliasID(outputInfo.OutputID, outputInfo.Output.(*iotago.AliasOutput))
}

func getAliasIDOtherOutputs(output iotago.Output) iotago.AliasID {
	var addressToCheck iotago.Address
	switch output.Type() {
	case iotago.OutputBasic:
		addressToCheck = output.(*iotago.BasicOutput).Ident()

	case iotago.OutputAlias:
		// TODO: chains can't own other alias outputs for now
		return iotago.AliasID{}

	case iotago.OutputFoundry:
		addressToCheck = output.(*iotago.FoundryOutput).Ident()

	case iotago.OutputNFT:
		addressToCheck = output.(*iotago.NFTOutput).Ident()

	default:
		panic(fmt.Errorf("%w: type %d", iotago.ErrUnknownOutputType, output.Type()))
	}

	if addressToCheck.Type() != iotago.AddressAlias {
		// output is not owned by an alias address => ignore it
		// TODO: what if we have nested ownerships? do we need to take care of that?
		return iotago.AliasID{}
	}

	return addressToCheck.(*iotago.AliasAddress).AliasID()
}

// filterAndSortAliasOutputs filters and groups all alias outputs by alias ID and then sorts them,
// because they could have been transitioned several times in the same milestone. applying the alias outputs to the consensus
// we need to apply them in correct order.
// chainsLock needs to be read locked outside
func filterAndSortAliasOutputs(chainsMap *shrinkingmap.ShrinkingMap[iotago.AliasID, *ncChain], ledgerUpdate *ledgerUpdate) (map[iotago.AliasID][]*isc.OutputInfo, map[iotago.OutputID]struct{}, error) {
	// filter and group "created alias outputs" by alias ID and also remember the tracked outputs
	trackedAliasOutputsCreatedMapByOutputID := make(map[iotago.OutputID]struct{})
	trackedAliasOutputsCreatedMapByAliasID := make(map[iotago.AliasID][]*isc.OutputInfo)
	for outputID := range ledgerUpdate.outputsCreatedMap {
		outputInfo := ledgerUpdate.outputsCreatedMap[outputID]

		aliasID := getAliasIDAliasOutput(outputInfo)
		if aliasID.Empty() {
			continue
		}

		// only allow tracked chains
		if !chainsMap.Has(aliasID) {
			continue
		}

		trackedAliasOutputsCreatedMapByOutputID[outputInfo.OutputID] = struct{}{}

		if _, exists := trackedAliasOutputsCreatedMapByAliasID[aliasID]; !exists {
			trackedAliasOutputsCreatedMapByAliasID[aliasID] = make([]*isc.OutputInfo, 0)
		}

		trackedAliasOutputsCreatedMapByAliasID[aliasID] = append(trackedAliasOutputsCreatedMapByAliasID[aliasID], outputInfo)
	}

	// create a map for faster lookups of output IDs that were spent by a transaction ID.
	// this is needed to figure out the correct ordering of alias outputs in case of governance transitions.
	trackedAliasOutputsConsumedMapByTransactionID := make(map[iotago.TransactionID]map[iotago.OutputID]struct{})
	for outputID := range ledgerUpdate.outputsConsumedMap {
		outputInfo := ledgerUpdate.outputsConsumedMap[outputID]

		aliasID := getAliasIDAliasOutput(outputInfo)
		if aliasID.Empty() {
			continue
		}

		// only allow tracked chains
		if !chainsMap.Has(aliasID) {
			continue
		}

		if _, exists := trackedAliasOutputsConsumedMapByTransactionID[outputInfo.TransactionIDSpent]; !exists {
			trackedAliasOutputsConsumedMapByTransactionID[outputInfo.TransactionIDSpent] = make(map[iotago.OutputID]struct{})
		}

		if _, exists := trackedAliasOutputsConsumedMapByTransactionID[outputInfo.TransactionIDSpent][outputInfo.OutputID]; !exists {
			trackedAliasOutputsConsumedMapByTransactionID[outputInfo.TransactionIDSpent][outputInfo.OutputID] = struct{}{}
		}
	}

	for aliasID := range trackedAliasOutputsCreatedMapByAliasID {
		if err := sortAliasOutputsOfChain(
			trackedAliasOutputsCreatedMapByAliasID[aliasID],
			trackedAliasOutputsConsumedMapByTransactionID,
		); err != nil {
			return nil, nil, err
		}
	}

	return trackedAliasOutputsCreatedMapByAliasID, trackedAliasOutputsCreatedMapByOutputID, nil
}

// chainsLock needs to be read locked
func filterOtherOutputs(
	chainsMap *shrinkingmap.ShrinkingMap[iotago.AliasID, *ncChain],
	outputsCreatedMap map[iotago.OutputID]*isc.OutputInfo,
	trackedAliasOutputsCreatedMapByOutputID map[iotago.OutputID]struct{},
) map[iotago.AliasID][]*isc.OutputInfo {
	otherOutputsCreatedByAliasID := make(map[iotago.AliasID][]*isc.OutputInfo)

	// we need to filter all other output types in case they were consumed in the same milestone.
	for outputID := range outputsCreatedMap {
		outputInfo := outputsCreatedMap[outputID]

		if _, exists := trackedAliasOutputsCreatedMapByOutputID[outputInfo.OutputID]; exists {
			// output will already be applied
			continue
		}

		if outputInfo.Consumed() {
			// output is not an important alias output that belongs to a chain,
			// and it was already consumed in the same milestone => ignore it
			continue
		}

		aliasID := getAliasIDOtherOutputs(outputInfo.Output)
		if aliasID.Empty() {
			continue
		}

		// allow only tracked chains
		if !chainsMap.Has(aliasID) {
			continue
		}

		if _, exists := otherOutputsCreatedByAliasID[aliasID]; !exists {
			otherOutputsCreatedByAliasID[aliasID] = make([]*isc.OutputInfo, 0)
		}

		// add the output to the tracked chain
		otherOutputsCreatedByAliasID[aliasID] = append(otherOutputsCreatedByAliasID[aliasID], outputInfo)
	}

	return otherOutputsCreatedByAliasID
}
