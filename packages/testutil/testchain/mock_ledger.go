package testchain

import (
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

type MockedLedger struct {
	latestOutputID                 *iotago.UTXOInput
	outputs                        map[iotago.UTXOInput]*iotago.AliasOutput
	txIDs                          map[iotago.TransactionID]bool
	publishTransactionAllowedFun   func(stateIndex uint32, tx *iotago.Transaction) bool
	pullLatestOutputAllowed        bool
	pullTxInclusionStateAllowedFun func(iotago.TransactionID) bool
	pullOutputByIDAllowedFun       func(*iotago.UTXOInput) bool
	pushOutputToNodesNeededFun     func(uint32, *iotago.Transaction, *iotago.UTXOInput, iotago.Output) bool
	stateOutputHandlerFuns         map[string]func(iotago.OutputID, iotago.Output)
	outputHandlerFuns              map[string]func(iotago.OutputID, iotago.Output)
	inclusionStateEvents           map[string]*events.Event
	mutex                          sync.RWMutex
	log                            *logger.Logger
}

func NewMockedLedger(chainID *iscp.ChainID, stateAddress iotago.Address, log *logger.Logger) *MockedLedger {
	originOutput := &iotago.AliasOutput{
		Amount:        iotago.TokenSupply,
		AliasID:       *chainID.AsAliasID(),
		StateMetadata: state.NewL1Commitment(state.OriginStateCommitment()).Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateAddress},
			&iotago.GovernorAddressUnlockCondition{Address: stateAddress},
		},
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: stateAddress,
			},
		},
	}
	outputID := getOriginOutputID()
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
	ret.SetPublishTransactionAllowed(true)
	ret.SetPullLatestOutputAllowed(true)
	ret.SetPullTxInclusionStateAllowed(true)
	ret.SetPullOutputByIDAllowed(true)
	ret.SetPushOutputToNodesNeeded(true)
	return ret
}

func (mlT *MockedLedger) Register(nodeID string, stateOutputHandler, outputHandler func(iotago.OutputID, iotago.Output)) {
	_, ok := mlT.outputHandlerFuns[nodeID]
	if ok {
		mlT.log.Panicf("Output handler for node %v already registered", nodeID)
	}
	mlT.stateOutputHandlerFuns[nodeID] = stateOutputHandler
	mlT.outputHandlerFuns[nodeID] = outputHandler
}

func (mlT *MockedLedger) Unregister(nodeID string) {
	delete(mlT.stateOutputHandlerFuns, nodeID)
	delete(mlT.outputHandlerFuns, nodeID)
}

func (mlT *MockedLedger) PublishTransaction(stateIndex uint32, tx *iotago.Transaction) error {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.log.Debugf("Publishing transaction for state %v", stateIndex)
	if mlT.publishTransactionAllowedFun(stateIndex, tx) {
		mlT.log.Debugf("Publishing transaction for state %v allowed, transaction has %v inputs, %v outputs, %v unlock blocks",
			stateIndex, len(tx.Essence.Inputs), len(tx.Essence.Outputs), len(tx.UnlockBlocks))
		txID, err := tx.ID()
		if err != nil {
			mlT.log.Panicf("Publishing transaction for state %v: cannot calculate transaction id: %v", stateIndex, err)
		}
		mlT.txIDs[*txID] = true
		for index, output := range tx.Essence.Outputs {
			aliasOutput, ok := output.(*iotago.AliasOutput)
			outputID := iotago.OutputIDFromTransactionIDAndIndex(*txID, uint16(index)).UTXOInput()
			mlT.log.Debugf("Publishing transaction for state %v: outputs[%v] has id %v", stateIndex, index, iscp.OID(outputID))
			if ok {
				mlT.log.Debugf("Publishing transaction for state %v: outputs[%v] is alias output", stateIndex, index)
				mlT.outputs[*outputID] = aliasOutput
				currentLatestOutput := mlT.getOutput(mlT.latestOutputID)
				if currentLatestOutput == nil || currentLatestOutput.StateIndex < aliasOutput.StateIndex {
					mlT.log.Debugf("Publishing transaction for state %v: outputs[%v] is newer than current newest output (%v -> %v)",
						stateIndex, index, currentLatestOutput.StateIndex, aliasOutput.StateIndex)
					mlT.latestOutputID = outputID
				}
			}
			if mlT.pushOutputToNodesNeededFun(stateIndex, tx, outputID, output) {
				mlT.log.Debugf("Publishing transaction for state %v: pushing it to nodes", stateIndex)
				for nodeID, handler := range mlT.stateOutputHandlerFuns {
					mlT.log.Debugf("Publishing transaction for state %v: pushing it to node %v", stateIndex, nodeID)
					go handler(outputID.ID(), output)
				}
			} else {
				mlT.log.Debugf("Publishing transaction for state %v: pushing it to nodes not needed", stateIndex)
			}
		}
		return nil
	}
	return fmt.Errorf("Publishing transaction for state %v not allowed", stateIndex)
}

func (mlT *MockedLedger) PullLatestOutput(nodeID string) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling latest output")
	if mlT.pullLatestOutputAllowed {
		mlT.log.Debugf("Pulling latest output allowed")
		output := mlT.getLatestOutput()
		mlT.log.Debugf("Pulling latest output: output with id %v pulled", iscp.OID(mlT.latestOutputID))
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

	mlT.log.Debugf("Pulling transaction inclusion state for ID %v", iscp.TxID(&txID))
	if mlT.pullTxInclusionStateAllowedFun(txID) {
		_, ok := mlT.txIDs[txID]
		var stateStr string
		if ok {
			stateStr = "included"
		} else {
			stateStr = "noTransaction"
		}
		mlT.log.Debugf("Pulling transaction inclusion state for ID %v: result is %v", iscp.TxID(&txID), stateStr)
		event, ok := mlT.inclusionStateEvents[nodeID]
		if ok {
			event.Trigger(txID, stateStr)
		} else {
			mlT.log.Panicf("Pulling transaction inclusion state for ID %v: no event for node id %v", iscp.TxID(&txID), nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling transaction inclusion state for ID %v not allowed", iscp.TxID(&txID))
	}
}

func (mlT *MockedLedger) PullOutputByID(nodeID string, outputID *iotago.UTXOInput) {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Pulling output by id %v", iscp.OID(outputID))
	if mlT.pullOutputByIDAllowedFun(outputID) {
		mlT.log.Debugf("Pulling output by id %v allowed", iscp.OID(outputID))
		output := mlT.getOutput(outputID)
		if output == nil {
			mlT.log.Warnf("Pulling output by id %v failed: output not found", iscp.OID(outputID))
			return
		}
		mlT.log.Debugf("Pulling output by id %v was successful", iscp.OID(outputID))
		handler, ok := mlT.stateOutputHandlerFuns[nodeID]
		if ok {
			go handler(outputID.ID(), output)
		} else {
			mlT.log.Panicf("Pulling output by id %v: no output handler for node id %v", iscp.OID(outputID), nodeID)
		}
	} else {
		mlT.log.Errorf("Pulling output by id %v not allowed", iscp.OID(outputID))
	}
}

func (mlT *MockedLedger) GetLatestOutput() *iscp.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting latest output")
	return iscp.NewAliasOutputWithID(mlT.getLatestOutput(), mlT.latestOutputID)
}

func (mlT *MockedLedger) getLatestOutput() *iotago.AliasOutput {
	output := mlT.getOutput(mlT.latestOutputID)
	if output == nil {
		mlT.log.Panicf("Latest output with id %v not found", iscp.OID(mlT.latestOutputID))
	}
	return output
}

func (mlT *MockedLedger) GetOutputByID(id *iotago.UTXOInput) *iotago.AliasOutput {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	mlT.log.Debugf("Getting output by ID %v", iscp.OID(id))
	return mlT.getOutput(id)
}

func (mlT *MockedLedger) getOutput(id *iotago.UTXOInput) *iotago.AliasOutput {
	output, ok := mlT.outputs[*id]
	if ok {
		return output
	}
	return nil
}

func (mlT *MockedLedger) AttachTxInclusionStateEvents(nodeID string, handler chain.NodeConnectionInclusionStateHandlerFun) (*events.Closure, error) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	closure := events.NewClosure(handler)
	event, ok := mlT.inclusionStateEvents[nodeID]
	if !ok {
		event = events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(chain.NodeConnectionInclusionStateHandlerFun)(params[0].(iotago.TransactionID), params[1].(string))
		})
		mlT.inclusionStateEvents[nodeID] = event
	}
	event.Attach(closure)
	return closure, nil
}

func (mlT *MockedLedger) DetachTxInclusionStateEvents(nodeID string, closure *events.Closure) error {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	event, ok := mlT.inclusionStateEvents[nodeID]
	if !ok {
		mlT.log.Panicf("Cannot dettach from event of node %v: no such event", nodeID)
	}
	event.Detach(closure)
	return nil
}

func (mlT *MockedLedger) SetPublishTransactionAllowed(flag bool) {
	mlT.SetPublishTransactionAllowedFun(func(uint32, *iotago.Transaction) bool { return flag })
}

func (mlT *MockedLedger) SetPublishTransactionAllowedFun(fun func(stateIndex uint32, tx *iotago.Transaction) bool) {
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
	mlT.SetPushOutputToNodesNeededFun(func(uint32, *iotago.Transaction, *iotago.UTXOInput, iotago.Output) bool { return flag })
}

func (mlT *MockedLedger) SetPushOutputToNodesNeededFun(fun func(state uint32, tx *iotago.Transaction, outputID *iotago.UTXOInput, output iotago.Output) bool) {
	mlT.mutex.Lock()
	defer mlT.mutex.Unlock()

	mlT.pushOutputToNodesNeededFun = fun
}

func getOriginOutputID() *iotago.UTXOInput {
	return &iotago.UTXOInput{}
}

func (mlT *MockedLedger) GetOriginOutput() *iscp.AliasOutputWithID {
	mlT.mutex.RLock()
	defer mlT.mutex.RUnlock()

	outputID := getOriginOutputID()
	output := mlT.getOutput(outputID)
	if output == nil {
		return nil
	}
	return iscp.NewAliasOutputWithID(output, outputID)
}
