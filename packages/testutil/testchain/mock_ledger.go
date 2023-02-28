package testchain

import (
	"errors"
	"sync"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
)

type MockedLedger struct {
	latestOutputID                 iotago.OutputID
	outputs                        map[iotago.OutputID]*iotago.AliasOutput
	txIDs                          map[iotago.TransactionID]bool
	publishTransactionAllowedFun   func(tx *iotago.Transaction) bool
	pullLatestOutputAllowed        bool
	pullTxInclusionStateAllowedFun func(iotago.TransactionID) bool
	pullOutputByIDAllowedFun       func(iotago.OutputID) bool
	pushOutputToNodesNeededFun     func(*iotago.Transaction, iotago.OutputID, iotago.Output) bool
	stateOutputHandlerFuns         map[string]func(iotago.OutputID, iotago.Output)
	outputHandlerFuns              map[string]func(iotago.OutputID, iotago.Output)
	inclusionStateEvents           map[string]*events.Event
	mutex                          sync.RWMutex
	log                            *logger.Logger
}

func NewMockedLedger(stateAddress iotago.Address, log *logger.Logger) (*MockedLedger, isc.ChainID) {
	originOutput := &iotago.AliasOutput{
		Amount:        tpkg.TestTokenSupply,
		StateMetadata: origin.L1Commitment(nil, 0).Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateAddress},
			&iotago.GovernorAddressUnlockCondition{Address: stateAddress},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: stateAddress,
			},
		},
	}
	outputID := getOriginOutputID()
	chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(outputID))
	originOutput.AliasID = chainID.AsAliasID() // NOTE: not very correct: origin output's AliasID should be empty; left here to make mocking transitions easier
	outputs := make(map[iotago.OutputID]*iotago.AliasOutput)
	outputs[outputID] = originOutput
	ret := &MockedLedger{
		latestOutputID:         outputID,
		outputs:                outputs,
		txIDs:                  make(map[iotago.TransactionID]bool),
		stateOutputHandlerFuns: make(map[string]func(iotago.OutputID, iotago.Output)),
		outputHandlerFuns:      make(map[string]func(iotago.OutputID, iotago.Output)),
		inclusionStateEvents:   make(map[string]*events.Event),
		log:                    log.Named("ml-" + chainID.String()[2:8]),
	}
	ret.SetPublishStateTransactionAllowed(true)
	ret.SetPublishGovernanceTransactionAllowed(true)
	ret.SetPullLatestOutputAllowed(true)
	ret.SetPullTxInclusionStateAllowed(true)
	ret.SetPullOutputByIDAllowed(true)
	ret.SetPushOutputToNodesNeeded(true)
	return ret, chainID
}

func (mlT *MockedLedger) Register(nodeID string, stateOutputHandler, outputHandler func(iotago.OutputID, iotago.Output)) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	_, ok := mlT.outputHandlerFuns[nodeID]
	if ok {
		mlT.log.Panicf("Output handler for node %v already registered", nodeID)
	}
	mlT.stateOutputHandlerFuns[nodeID] = stateOutputHandler
	mlT.outputHandlerFuns[nodeID] = outputHandler
}

func (mlT *MockedLedger) Unregister(nodeID string) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	delete(mlT.stateOutputHandlerFuns, nodeID)
	delete(mlT.outputHandlerFuns, nodeID)
}

func (mlT *MockedLedger) PublishTransaction(tx *iotago.Transaction) error {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	if mlT.publishTransactionAllowedFun(tx) {
		mlT.log.Debugf("Publishing transaction allowed, transaction has %v inputs, %v outputs, %v unlock blocks",
			len(tx.Essence.Inputs), len(tx.Essence.Outputs), len(tx.Unlocks))
		txID, err := tx.ID()
		if err != nil {
			mlT.log.Panicf("Publishing transaction: cannot calculate transaction id: %v", err)
		}
		mlT.log.Debugf("Publishing transaction: transaction id is %s", txID.ToHex())
		mlT.txIDs[txID] = true
		for index, output := range tx.Essence.Outputs {
			aliasOutput, ok := output.(*iotago.AliasOutput)
			outputID := iotago.OutputIDFromTransactionIDAndIndex(txID, uint16(index))
			mlT.log.Debugf("Publishing transaction: outputs[%v] has id %v", index, outputID.ToHex())
			if ok {
				mlT.log.Debugf("Publishing transaction: outputs[%v] is alias output", index)
				mlT.outputs[outputID] = aliasOutput
				currentLatestAliasOutput := mlT.getAliasOutput(mlT.latestOutputID)
				if currentLatestAliasOutput == nil || currentLatestAliasOutput.StateIndex < aliasOutput.StateIndex {
					mlT.log.Debugf("Publishing transaction: outputs[%v] is newer than current newest output (%v -> %v)",
						index, currentLatestAliasOutput.StateIndex, aliasOutput.StateIndex)
					mlT.latestOutputID = outputID
				}
			}
			if mlT.pushOutputToNodesNeededFun(tx, outputID, output) {
				mlT.log.Debugf("Publishing transaction: pushing it to nodes")
				for nodeID, handler := range mlT.stateOutputHandlerFuns {
					mlT.log.Debugf("Publishing transaction: pushing it to node %v", nodeID)
					go handler(outputID, output)
				}
			} else {
				mlT.log.Debugf("Publishing transaction: pushing it to nodes not needed")
			}
		}
		return nil
	}
	return errors.New("publishing transaction not allowed")
}

func (mlT *MockedLedger) PullLatestOutput(nodeID string) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling latest output")
	if mlT.pullLatestOutputAllowed {
		mlT.log.Debugf("Pulling latest output allowed")
		output := mlT.getLatestOutput()
		mlT.log.Debugf("Pulling latest output: output with id %v pulled", mlT.latestOutputID.ToHex())
		handler, ok := mlT.stateOutputHandlerFuns[nodeID]
		if ok {
			go handler(mlT.latestOutputID, output)
		} else {
			mlT.log.Panicf("Pulling latest output: no output handler for node id %v", nodeID)
		}
	} else {
		mlT.log.Error("Pulling latest output not allowed")
	}
}

func (mlT *MockedLedger) PullTxInclusionState(nodeID string, txID iotago.TransactionID) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	txIDHex := txID.ToHex()
	mlT.log.Debugf("Pulling transaction inclusion state for ID %v", txIDHex)
	if mlT.pullTxInclusionStateAllowedFun(txID) {
		_, ok := mlT.txIDs[txID]
		var stateStr string
		if ok {
			stateStr = "included"
		} else {
			stateStr = "noTransaction"
		}
		mlT.log.Debugf("Pulling transaction inclusion state for ID %v: result is %v", txIDHex, stateStr)
		event, ok := mlT.inclusionStateEvents[nodeID]
		if ok {
			event.Trigger(txID, stateStr)
		} else {
			mlT.log.Panicf("Pulling transaction inclusion state for ID %v: no event for node id %v", txIDHex, nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling transaction inclusion state for ID %v not allowed", txIDHex)
	}
}

func (mlT *MockedLedger) PullStateOutputByID(nodeID string, outputID iotago.OutputID) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	outputIDHex := outputID.ToHex()

	mlT.log.Debugf("Pulling output by id %v", outputIDHex)
	if mlT.pullOutputByIDAllowedFun(outputID) {
		mlT.log.Debugf("Pulling output by id %v allowed", outputIDHex)
		aliasOutput := mlT.getAliasOutput(outputID)
		if aliasOutput == nil {
			mlT.log.Warnf("Pulling output by id %v failed: output not found", outputIDHex)
			return
		}
		mlT.log.Debugf("Pulling output by id %v was successful", outputIDHex)
		handler, ok := mlT.stateOutputHandlerFuns[nodeID]
		if ok {
			go handler(outputID, aliasOutput)
		} else {
			mlT.log.Panicf("Pulling output by id %v: no output handler for node id %v", outputIDHex, nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling output by id %v not allowed", outputIDHex)
	}
}

func (mlT *MockedLedger) GetLatestOutput() *isc.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting latest output")
	return isc.NewAliasOutputWithID(mlT.getLatestOutput(), mlT.latestOutputID)
}

func (mlT *MockedLedger) getLatestOutput() *iotago.AliasOutput {
	aliasOutput := mlT.getAliasOutput(mlT.latestOutputID)
	if aliasOutput == nil {
		mlT.log.Panicf("Latest output with id %v not found", mlT.latestOutputID.ToHex())
	}
	return aliasOutput
}

func (mlT *MockedLedger) GetAliasOutputByID(outputID iotago.OutputID) *iotago.AliasOutput {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting alias output by ID %v", outputID.ToHex())
	return mlT.getAliasOutput(outputID)
}

func (mlT *MockedLedger) getAliasOutput(outputID iotago.OutputID) *iotago.AliasOutput {
	output, ok := mlT.outputs[outputID]
	if ok {
		return output
	}
	return nil
}

func (mlT *MockedLedger) SetPublishStateTransactionAllowed(flag bool) {
	mlT.SetPublishStateTransactionAllowedFun(func(*iotago.Transaction) bool { return flag })
}

func (mlT *MockedLedger) SetPublishStateTransactionAllowedFun(fun func(tx *iotago.Transaction) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.publishTransactionAllowedFun = fun
}

func (mlT *MockedLedger) SetPublishGovernanceTransactionAllowed(flag bool) {
	mlT.SetPublishGovernanceTransactionAllowedFun(func(*iotago.Transaction) bool { return flag })
}

func (mlT *MockedLedger) SetPublishGovernanceTransactionAllowedFun(fun func(tx *iotago.Transaction) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.publishTransactionAllowedFun = fun
}

func (mlT *MockedLedger) SetPullLatestOutputAllowed(flag bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pullLatestOutputAllowed = flag
}

func (mlT *MockedLedger) SetPullTxInclusionStateAllowed(flag bool) {
	mlT.SetPullTxInclusionStateAllowedFun(func(iotago.TransactionID) bool { return flag })
}

func (mlT *MockedLedger) SetPullTxInclusionStateAllowedFun(fun func(txID iotago.TransactionID) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pullTxInclusionStateAllowedFun = fun
}

func (mlT *MockedLedger) SetPullOutputByIDAllowed(flag bool) {
	mlT.SetPullOutputByIDAllowedFun(func(iotago.OutputID) bool { return flag })
}

func (mlT *MockedLedger) SetPullOutputByIDAllowedFun(fun func(outputID iotago.OutputID) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pullOutputByIDAllowedFun = fun
}

func (mlT *MockedLedger) SetPushOutputToNodesNeeded(flag bool) {
	mlT.SetPushOutputToNodesNeededFun(func(*iotago.Transaction, iotago.OutputID, iotago.Output) bool { return flag })
}

func (mlT *MockedLedger) SetPushOutputToNodesNeededFun(fun func(tx *iotago.Transaction, outputID iotago.OutputID, output iotago.Output) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pushOutputToNodesNeededFun = fun
}

func getOriginOutputID() iotago.OutputID {
	return iotago.OutputID{}
}

func (mlT *MockedLedger) GetOriginOutput() *isc.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	outputID := getOriginOutputID()
	aliasOutput := mlT.getAliasOutput(outputID)
	if aliasOutput == nil {
		return nil
	}
	return isc.NewAliasOutputWithID(aliasOutput, outputID)
}
