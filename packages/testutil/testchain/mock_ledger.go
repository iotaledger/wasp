package testchain

import (
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type MockedLedger struct {
	latestOutputID                 *iotago.UTXOInput
	outputs                        map[iotago.UTXOInput]*iotago.AliasOutput
	txIDs                          map[iotago.TransactionID]bool
	publishTransactionAllowedFun   func(tx *iotago.Transaction) bool
	pullLatestOutputAllowed        bool
	pullTxInclusionStateAllowedFun func(iotago.TransactionID) bool
	pullOutputByIDAllowedFun       func(*iotago.UTXOInput) bool
	pushOutputToNodesNeededFun     func(*iotago.Transaction, *iotago.UTXOInput, iotago.Output) bool
	stateOutputHandlerFuns         map[string]func(iotago.OutputID, iotago.Output)
	outputHandlerFuns              map[string]func(iotago.OutputID, iotago.Output)
	inclusionStateEvents           map[string]*events.Event
	mutex                          sync.RWMutex
	log                            *logger.Logger
}

func NewMockedLedger(stateAddress iotago.Address, log *logger.Logger) (*MockedLedger, *isc.ChainID) {
	originOutput := &iotago.AliasOutput{
		Amount:        tpkg.TestTokenSupply,
		StateMetadata: state.OriginL1Commitment().Bytes(),
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
	chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(outputID.ID()))
	originOutput.AliasID = *chainID.AsAliasID() // NOTE: not very correct: origin output's AliasID should be empty; left here to make mocking transitions easier
	outputs := make(map[iotago.UTXOInput]*iotago.AliasOutput)
	outputs[*outputID] = originOutput
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
	return ret, &chainID
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
		mlT.log.Debugf("Publishing transaction: transaction id is %s", isc.TxID(txID))
		mlT.txIDs[txID] = true
		for index, output := range tx.Essence.Outputs {
			aliasOutput, ok := output.(*iotago.AliasOutput)
			outputID := iotago.OutputIDFromTransactionIDAndIndex(txID, uint16(index)).UTXOInput()
			mlT.log.Debugf("Publishing transaction: outputs[%v] has id %v", index, isc.OID(outputID))
			if ok {
				mlT.log.Debugf("Publishing transaction: outputs[%v] is alias output", index)
				mlT.outputs[*outputID] = aliasOutput
				currentLatestOutput := mlT.getOutput(mlT.latestOutputID)
				if currentLatestOutput == nil || currentLatestOutput.StateIndex < aliasOutput.StateIndex {
					mlT.log.Debugf("Publishing transaction: outputs[%v] is newer than current newest output (%v -> %v)",
						index, currentLatestOutput.StateIndex, aliasOutput.StateIndex)
					mlT.latestOutputID = outputID
				}
			}
			if mlT.pushOutputToNodesNeededFun(tx, outputID, output) {
				mlT.log.Debugf("Publishing transaction: pushing it to nodes")
				for nodeID, handler := range mlT.stateOutputHandlerFuns {
					mlT.log.Debugf("Publishing transaction: pushing it to node %v", nodeID)
					go handler(outputID.ID(), output)
				}
			} else {
				mlT.log.Debugf("Publishing transaction: pushing it to nodes not needed")
			}
		}
		return nil
	}
	return fmt.Errorf("Publishing transaction not allowed")
}

func (mlT *MockedLedger) PullLatestOutput(nodeID string) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling latest output")
	if mlT.pullLatestOutputAllowed {
		mlT.log.Debugf("Pulling latest output allowed")
		output := mlT.getLatestOutput()
		mlT.log.Debugf("Pulling latest output: output with id %v pulled", isc.OID(mlT.latestOutputID))
		handler, ok := mlT.stateOutputHandlerFuns[nodeID]
		if ok {
			go handler(mlT.latestOutputID.ID(), output)
		} else {
			mlT.log.Panicf("Pulling latest output: no output handler for node id %v", nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling latest output not allowed")
	}
}

func (mlT *MockedLedger) PullTxInclusionState(nodeID string, txID iotago.TransactionID) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling transaction inclusion state for ID %v", isc.TxID(txID))
	if mlT.pullTxInclusionStateAllowedFun(txID) {
		_, ok := mlT.txIDs[txID]
		var stateStr string
		if ok {
			stateStr = "included"
		} else {
			stateStr = "noTransaction"
		}
		mlT.log.Debugf("Pulling transaction inclusion state for ID %v: result is %v", isc.TxID(txID), stateStr)
		event, ok := mlT.inclusionStateEvents[nodeID]
		if ok {
			event.Trigger(txID, stateStr)
		} else {
			mlT.log.Panicf("Pulling transaction inclusion state for ID %v: no event for node id %v", isc.TxID(txID), nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling transaction inclusion state for ID %v not allowed", isc.TxID(txID))
	}
}

func (mlT *MockedLedger) PullStateOutputByID(nodeID string, outputID *iotago.UTXOInput) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling output by id %v", isc.OID(outputID))
	if mlT.pullOutputByIDAllowedFun(outputID) {
		mlT.log.Debugf("Pulling output by id %v allowed", isc.OID(outputID))
		output := mlT.getOutput(outputID)
		if output == nil {
			mlT.log.Warnf("Pulling output by id %v failed: output not found", isc.OID(outputID))
			return
		}
		mlT.log.Debugf("Pulling output by id %v was successful", isc.OID(outputID))
		handler, ok := mlT.stateOutputHandlerFuns[nodeID]
		if ok {
			go handler(outputID.ID(), output)
		} else {
			mlT.log.Panicf("Pulling output by id %v: no output handler for node id %v", isc.OID(outputID), nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling output by id %v not allowed", isc.OID(outputID))
	}
}

func (mlT *MockedLedger) GetLatestOutput() *isc.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting latest output")
	return isc.NewAliasOutputWithID(mlT.getLatestOutput(), mlT.latestOutputID)
}

func (mlT *MockedLedger) getLatestOutput() *iotago.AliasOutput {
	output := mlT.getOutput(mlT.latestOutputID)
	if output == nil {
		mlT.log.Panicf("Latest output with id %v not found", isc.OID(mlT.latestOutputID))
	}
	return output
}

func (mlT *MockedLedger) GetOutputByID(id *iotago.UTXOInput) *iotago.AliasOutput {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting output by ID %v", isc.OID(id))
	return mlT.getOutput(id)
}

func (mlT *MockedLedger) getOutput(id *iotago.UTXOInput) *iotago.AliasOutput {
	output, ok := mlT.outputs[*id]
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
	mlT.SetPullOutputByIDAllowedFun(func(*iotago.UTXOInput) bool { return flag })
}

func (mlT *MockedLedger) SetPullOutputByIDAllowedFun(fun func(outputID *iotago.UTXOInput) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pullOutputByIDAllowedFun = fun
}

func (mlT *MockedLedger) SetPushOutputToNodesNeeded(flag bool) {
	mlT.SetPushOutputToNodesNeededFun(func(*iotago.Transaction, *iotago.UTXOInput, iotago.Output) bool { return flag })
}

func (mlT *MockedLedger) SetPushOutputToNodesNeededFun(fun func(tx *iotago.Transaction, outputID *iotago.UTXOInput, output iotago.Output) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pushOutputToNodesNeededFun = fun
}

func getOriginOutputID() *iotago.UTXOInput {
	return &iotago.UTXOInput{}
}

func (mlT *MockedLedger) GetOriginOutput() *isc.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	outputID := getOriginOutputID()
	output := mlT.getOutput(outputID)
	if output == nil {
		return nil
	}
	return isc.NewAliasOutputWithID(output, outputID)
}
