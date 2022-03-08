package testchain

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
)

type MockedLedger struct {
	log                             *logger.Logger
	latestOutputID                  *iotago.UTXOInput
	outputs                         map[iotago.UTXOInput]*iotago.AliasOutput
	pushTransactionToNodesNeededFun func(tx *iotago.Transaction) bool
	nodes                           []*MockedNodeConn
	mutex                           sync.RWMutex
}

func NewMockedLedger(originOutput *iotago.AliasOutput, log *logger.Logger) *MockedLedger {
	outputID := iotago.UTXOInput{}
	outputs := make(map[iotago.UTXOInput]*iotago.AliasOutput)
	outputs[outputID] = originOutput
	ret := &MockedLedger{
		log:            log.Named("ledger"),
		latestOutputID: &outputID,
		outputs:        outputs,
		nodes:          make([]*MockedNodeConn, 0),
	}
	ret.SetPushTransactionToNodesNeeded(true)
	return ret
}

func (m *MockedLedger) pullState() *iscp.AliasOutputWithID {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.log.Debugf("Pulling state")
	output := m.getOutput(m.latestOutputID)
	if output != nil {
		m.log.Debugf("Output with id %v pulled", iscp.OID(m.latestOutputID))
		return iscp.NewAliasOutputWithID(output, m.latestOutputID)
	}
	m.log.Debugf("Output with id %v not found", iscp.OID(m.latestOutputID))
	return nil
}

func (m *MockedLedger) PullConfirmedOutput(id *iotago.UTXOInput) *iotago.AliasOutput {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.log.Debugf("Pulling confirmed output with id %v", iscp.OID(id))
	output := m.getOutput(id)
	if output != nil {
		m.log.Debugf("Confirmed output with id %v found", iscp.OID(id))
		return output
	}
	m.log.Debugf("Confirmed output with id %v not found", iscp.OID(id))
	return nil
}

func (m *MockedLedger) getOutput(id *iotago.UTXOInput) *iotago.AliasOutput {
	output, ok := m.outputs[*id]
	if ok {
		return output
	}
	return nil
}

func (m *MockedLedger) receiveTx(tx *iotago.Transaction) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.log.Debugf("Transaction received: %v inputs, %v outputs, %v unlock blocks",
		len(tx.Essence.Inputs), len(tx.Essence.Outputs), len(tx.UnlockBlocks))
	for index, output := range tx.Essence.Outputs {
		aliasOutput, ok := output.(*iotago.AliasOutput)
		if ok {
			txID, err := tx.ID()
			if err != nil {
				panic(err)
			}
			outputID := iotago.OutputIDFromTransactionIDAndIndex(*txID, uint16(index)).UTXOInput()
			m.log.Debugf("Transaction received: outputs[%v] is alias output with id %v", index, iscp.OID(outputID))
			m.outputs[*outputID] = aliasOutput
			currentLatestOutput := m.getOutput(m.latestOutputID)
			if currentLatestOutput == nil || currentLatestOutput.StateIndex < aliasOutput.StateIndex {
				m.log.Debugf("Transaction received: outputs[%v] is newer than current newest output (%v -> %v)",
					index, aliasOutput.StateIndex, currentLatestOutput.StateIndex)
				m.latestOutputID = outputID
				if m.pushTransactionToNodesNeededFun(tx) {
					for _, node := range m.nodes {
						m.log.Debugf("Transaction received: pusshing it to node %v", node.ID())
						go node.handleTransactionFun(tx)
					}
				}
			}
		}
	}
}

func (m *MockedLedger) SetPushTransactionToNodesNeeded(needed bool) {
	m.SetPushTransactionToNodesNeededFun(func(*iotago.Transaction) bool { return needed })
}

func (m *MockedLedger) SetPushTransactionToNodesNeededFun(fun func(tx *iotago.Transaction) bool) {
	m.pushTransactionToNodesNeededFun = fun
}

func (m *MockedLedger) addNode(node *MockedNodeConn) {
	m.log.Debugf("Adding node %v", node.ID())
	m.nodes = append(m.nodes, node)
}
