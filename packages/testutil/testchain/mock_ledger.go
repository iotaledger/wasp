package testchain

import (
	"sync"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
)

type MockedLedger struct {
	latestOutputID                  *iotago.UTXOInput
	outputs                         map[iotago.UTXOInput]*iotago.AliasOutput
	pushTransactionToNodesNeededFun func(tx *iotago.Transaction) bool
	nodes                           []*MockedNodeConn
	mutex                           sync.RWMutex
}

func NewMockedLedger(originOutput *iotago.AliasOutput) *MockedLedger {
	outputID := iotago.UTXOInput{}
	outputs := make(map[iotago.UTXOInput]*iotago.AliasOutput)
	outputs[outputID] = originOutput
	return &MockedLedger{
		latestOutputID:                  &outputID,
		outputs:                         outputs,
		pushTransactionToNodesNeededFun: func(*iotago.Transaction) bool { return true },
		nodes:                           make([]*MockedNodeConn, 0),
	}
}

func (m *MockedLedger) pullState() *iscp.AliasOutputWithID {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	output := m.getOutput(m.latestOutputID)
	if output != nil {
		return iscp.NewAliasOutputWithID(output, m.latestOutputID)
	}
	return nil
}

func (m *MockedLedger) pullConfirmedOutput(id *iotago.UTXOInput) *iotago.AliasOutput {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	output := m.getOutput(id)
	if output != nil {
		return output
	}
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
	for index, output := range tx.Essence.Outputs {
		aliasOutput, ok := output.(*iotago.AliasOutput)
		if ok {
			txID, err := tx.ID()
			if err != nil {
				panic(err)
			}
			outputID := iotago.OutputIDFromTransactionIDAndIndex(*txID, uint16(index)).UTXOInput()
			m.outputs[*outputID] = aliasOutput
			currentLatestOutput := m.pullState()
			if currentLatestOutput == nil || currentLatestOutput.GetAliasOutput().StateIndex < aliasOutput.StateIndex {
				m.latestOutputID = outputID
				if m.pushTransactionToNodesNeededFun(tx) {
					for _, node := range m.nodes {
						node.handleTransactionFun(tx)
					}
				}
			}
		}
	}
}

func (m *MockedLedger) addNode(node *MockedNodeConn) {
	m.nodes = append(m.nodes, node)
}
