// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconnmetrics

import (
	"math/big"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestRegister(t *testing.T) {
	chainID1 := isc.RandomChainID()
	chainID2 := isc.RandomChainID()
	chainID3 := isc.RandomChainID()
	ncm := New()

	require.Equal(t, []isc.ChainID{}, ncm.GetRegistered())

	ncm.SetRegistered(chainID1)
	require.Equal(t, []isc.ChainID{chainID1}, ncm.GetRegistered())

	ncm.SetRegistered(chainID2)
	registered := ncm.GetRegistered()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)

	ncm.SetUnregistered(chainID1)
	require.Equal(t, []isc.ChainID{chainID2}, ncm.GetRegistered())

	ncm.SetRegistered(chainID3)
	registered = ncm.GetRegistered()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID2)
	require.Contains(t, registered, chainID3)

	ncm.SetUnregistered(chainID3)
	require.Equal(t, []isc.ChainID{chainID2}, ncm.GetRegistered())

	ncm.SetRegistered(chainID1)
	registered = ncm.GetRegistered()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)

	ncm.SetRegistered(chainID3)
	registered = ncm.GetRegistered()
	require.Equal(t, 3, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)
	require.Contains(t, registered, chainID3)
}

func createOnLedgerRequest() isc.OnLedgerRequest {
	requestMetadata := &isc.RequestMetadata{
		SenderContract: isc.Hn("sender_contract"),
		TargetContract: isc.Hn("target_contract"),
		EntryPoint:     isc.Hn("entrypoint"),
		Allowance:      isc.NewAllowanceBaseTokens(1),
		GasBudget:      1000,
	}

	outputOn := &iotago.BasicOutput{
		Amount: 123,
		NativeTokens: iotago.NativeTokens{
			&iotago.NativeToken{
				ID:     [iotago.NativeTokenIDLength]byte{1},
				Amount: big.NewInt(100),
			},
		},
		Features: iotago.Features{
			&iotago.MetadataFeature{Data: requestMetadata.Bytes()},
			&iotago.SenderFeature{Address: tpkg.RandAliasAddress()},
		},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
		},
	}

	onLedgerRequest1, _ := isc.OnLedgerFromUTXO(outputOn, iotago.OutputID{})
	return onLedgerRequest1
}

func TestMessageMetrics(t *testing.T) {
	ncm := New()
	cncm1 := ncm.NewMessagesMetrics(isc.RandomChainID())
	cncm2 := ncm.NewMessagesMetrics(isc.RandomChainID())
	ncm.Register(prometheus.NewRegistry())

	// IN State output
	outputID1 := &InStateOutput{OutputID: iotago.OutputID{1}}
	outputID2 := &InStateOutput{OutputID: iotago.OutputID{2}}
	outputID3 := &InStateOutput{OutputID: iotago.OutputID{3}}

	cncm1.GetInStateOutput().CountLastMessage(outputID1)
	cncm1.GetInStateOutput().CountLastMessage(outputID2)
	cncm1.GetInStateOutput().CountLastMessage(outputID3)

	checkMetricsValues(t, 3, outputID3, cncm1.GetInStateOutput())
	checkMetricsValues(t, 0, new(InStateOutput), cncm2.GetInStateOutput())
	checkMetricsValues(t, 3, outputID3, ncm.GetInStateOutput())

	// IN Alias output
	aliasOutput1 := &iotago.AliasOutput{StateIndex: 1}
	aliasOutput2 := &iotago.AliasOutput{StateIndex: 2}
	aliasOutput3 := &iotago.AliasOutput{StateIndex: 3}

	ncm.GetInAliasOutput().CountLastMessage(aliasOutput1)
	cncm1.GetInAliasOutput().CountLastMessage(aliasOutput2)
	cncm1.GetInAliasOutput().CountLastMessage(aliasOutput3)

	checkMetricsValues(t, 2, aliasOutput3, cncm1.GetInAliasOutput())
	checkMetricsValues(t, 0, new(iotago.AliasOutput), cncm2.GetInAliasOutput())
	checkMetricsValues(t, 3, aliasOutput3, ncm.GetInAliasOutput())

	// IN Output
	inOutput1 := &InOutput{OutputID: iotago.OutputID{1}}
	inOutput2 := &InOutput{OutputID: iotago.OutputID{2}}
	inOutput3 := &InOutput{OutputID: iotago.OutputID{3}}

	cncm1.GetInOutput().CountLastMessage(inOutput1)
	cncm2.GetInOutput().CountLastMessage(inOutput2)
	ncm.GetInOutput().CountLastMessage(inOutput3)

	checkMetricsValues(t, 1, inOutput1, cncm1.GetInOutput())
	checkMetricsValues(t, 1, inOutput2, cncm2.GetInOutput())
	checkMetricsValues(t, 3, inOutput3, ncm.GetInOutput())

	// IN Transaction inclusion state
	txInclusionState1 := &TxInclusionStateMsg{TxID: iotago.TransactionID{1}}
	txInclusionState2 := &TxInclusionStateMsg{TxID: iotago.TransactionID{2}}
	txInclusionState3 := &TxInclusionStateMsg{TxID: iotago.TransactionID{3}}

	cncm1.GetInTxInclusionState().CountLastMessage(txInclusionState1)
	cncm1.GetInTxInclusionState().CountLastMessage(txInclusionState2)
	cncm2.GetInTxInclusionState().CountLastMessage(txInclusionState3)

	checkMetricsValues(t, 2, txInclusionState2, cncm1.GetInTxInclusionState())
	checkMetricsValues(t, 1, txInclusionState3, cncm2.GetInTxInclusionState())
	checkMetricsValues(t, 3, txInclusionState3, ncm.GetInTxInclusionState())

	// IN On ledger request

	onLedgerRequest1 := createOnLedgerRequest()
	onLedgerRequest2 := createOnLedgerRequest()
	onLedgerRequest3 := createOnLedgerRequest()

	cncm1.GetInOnLedgerRequest().CountLastMessage(onLedgerRequest1)
	cncm2.GetInOnLedgerRequest().CountLastMessage(onLedgerRequest2)
	cncm1.GetInOnLedgerRequest().CountLastMessage(onLedgerRequest3)

	checkMetricsValues(t, 2, onLedgerRequest3, cncm1.GetInOnLedgerRequest())
	checkMetricsValues(t, 1, onLedgerRequest2, cncm2.GetInOnLedgerRequest())
	checkMetricsValues(t, 3, onLedgerRequest3, ncm.GetInOnLedgerRequest())

	// OUT Publish state transaction
	stateTransaction1 := &StateTransaction{StateIndex: 1}
	stateTransaction2 := &StateTransaction{StateIndex: 1}
	stateTransaction3 := &StateTransaction{StateIndex: 1}

	cncm1.GetOutPublishStateTransaction().CountLastMessage(stateTransaction1)
	cncm2.GetOutPublishStateTransaction().CountLastMessage(stateTransaction2)
	cncm2.GetOutPublishStateTransaction().CountLastMessage(stateTransaction3)

	checkMetricsValues(t, 1, stateTransaction1, cncm1.GetOutPublishStateTransaction())
	checkMetricsValues(t, 2, stateTransaction3, cncm2.GetOutPublishStateTransaction())
	checkMetricsValues(t, 3, stateTransaction3, ncm.GetOutPublishStateTransaction())

	// OUT Publish governance transaction
	publishStateTransaction1 := &iotago.Transaction{
		Essence: nil,
		Unlocks: nil,
	}
	publishStateTransaction2 := &iotago.Transaction{
		Essence: nil,
		Unlocks: nil,
	}
	publishStateTransaction3 := &iotago.Transaction{
		Essence: nil,
		Unlocks: nil,
	}

	cncm2.GetOutPublishGovernanceTransaction().CountLastMessage(publishStateTransaction1)
	cncm2.GetOutPublishGovernanceTransaction().CountLastMessage(publishStateTransaction2)
	cncm1.GetOutPublishGovernanceTransaction().CountLastMessage(publishStateTransaction3)

	checkMetricsValues(t, 1, publishStateTransaction3, cncm1.GetOutPublishGovernanceTransaction())
	checkMetricsValues(t, 2, publishStateTransaction2, cncm2.GetOutPublishGovernanceTransaction())
	checkMetricsValues(t, 3, publishStateTransaction3, ncm.GetOutPublishGovernanceTransaction())

	// OUT Pull latest output
	ncm.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput1")
	cncm1.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput2")
	cncm2.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput3")

	checkMetricsValues(t, 1, "OutPullLatestOutput2", cncm1.GetOutPullLatestOutput())
	checkMetricsValues(t, 1, "OutPullLatestOutput3", cncm2.GetOutPullLatestOutput())
	checkMetricsValues(t, 3, "OutPullLatestOutput3", ncm.GetOutPullLatestOutput())

	// OUT Pull transaction inclusion state
	transactionID1 := iotago.TransactionID{1}
	transactionID2 := iotago.TransactionID{2}
	transactionID3 := iotago.TransactionID{3}

	cncm1.GetOutPullTxInclusionState().CountLastMessage(transactionID1)
	ncm.GetOutPullTxInclusionState().CountLastMessage(transactionID2)
	cncm2.GetOutPullTxInclusionState().CountLastMessage(transactionID3)

	checkMetricsValues(t, 1, transactionID1, cncm1.GetOutPullTxInclusionState())
	checkMetricsValues(t, 1, transactionID3, cncm2.GetOutPullTxInclusionState())
	checkMetricsValues(t, 3, transactionID3, ncm.GetOutPullTxInclusionState())

	// OUT Pull output by ID
	utxoInput1 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}
	utxoInput2 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}
	utxoInput3 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}

	cncm1.GetOutPullOutputByID().CountLastMessage(utxoInput1.ID())
	cncm1.GetOutPullOutputByID().CountLastMessage(utxoInput2.ID())
	cncm1.GetOutPullOutputByID().CountLastMessage(utxoInput3.ID())

	checkMetricsValues(t, 3, utxoInput3.ID(), cncm1.GetOutPullOutputByID())
	checkMetricsValues(t, 0, iotago.OutputID{}, cncm2.GetOutPullOutputByID())
	checkMetricsValues(t, 3, utxoInput3.ID(), ncm.GetOutPullOutputByID())

	// IN Milestone
	milestoneInfo1 := &nodeclient.MilestoneInfo{Index: 0}
	milestoneInfo2 := &nodeclient.MilestoneInfo{Index: 0}

	ncm.GetInMilestone().CountLastMessage(milestoneInfo1)
	ncm.GetInMilestone().CountLastMessage(milestoneInfo2)

	checkMetricsValues(t, 2, milestoneInfo2, ncm.GetInMilestone())
}

func checkMetricsValues[T any, V any](t *testing.T, expectedTotal uint32, expectedLastMessage V, metrics NodeConnectionMessageMetrics[T]) {
	require.Equal(t, expectedTotal, metrics.GetMessageTotal())
	if expectedTotal == 0 {
		var zeroValue V
		require.Equal(t, zeroValue, metrics.GetLastMessage())
	} else {
		require.Equal(t, expectedLastMessage, metrics.GetLastMessage())
	}
}
