// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"math/big"
	"testing"

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
	ncm := NewChainMetricsProvider()

	require.Equal(t, []isc.ChainID{}, ncm.RegisteredChains())

	ncm.RegisterChain(chainID1)
	require.Equal(t, []isc.ChainID{chainID1}, ncm.RegisteredChains())

	ncm.RegisterChain(chainID2)
	registered := ncm.RegisteredChains()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)

	ncm.UnregisterChain(chainID1)
	require.Equal(t, []isc.ChainID{chainID2}, ncm.RegisteredChains())

	ncm.RegisterChain(chainID3)
	registered = ncm.RegisteredChains()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID2)
	require.Contains(t, registered, chainID3)

	ncm.UnregisterChain(chainID3)
	require.Equal(t, []isc.ChainID{chainID2}, ncm.RegisteredChains())

	ncm.RegisterChain(chainID1)
	registered = ncm.RegisteredChains()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)

	ncm.RegisterChain(chainID3)
	registered = ncm.RegisteredChains()
	require.Equal(t, 3, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)
	require.Contains(t, registered, chainID3)
}

func createOnLedgerRequest() isc.OnLedgerRequest {
	requestMetadata := &isc.RequestMetadata{
		SenderContract: isc.ContractIdentityFromHname(isc.Hn("sender_contract")),
		TargetContract: isc.Hn("target_contract"),
		EntryPoint:     isc.Hn("entrypoint"),
		Allowance:      isc.NewAssetsBaseTokens(1),
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
	p := NewChainMetricsProvider()
	ncm := p.Message
	cncm1 := p.GetChainMetrics(isc.RandomChainID()).Message
	cncm2 := p.GetChainMetrics(isc.RandomChainID()).Message

	// IN State output
	outputID1 := &InStateOutput{OutputID: iotago.OutputID{1}}
	outputID2 := &InStateOutput{OutputID: iotago.OutputID{2}}
	outputID3 := &InStateOutput{OutputID: iotago.OutputID{3}}

	cncm1.InStateOutput().IncMessages(outputID1)
	cncm1.InStateOutput().IncMessages(outputID2)
	cncm1.InStateOutput().IncMessages(outputID3)

	checkMetricsValues(t, 3, outputID3, cncm1.InStateOutput())
	checkMetricsValues(t, 0, new(InStateOutput), cncm2.InStateOutput())
	checkMetricsValues(t, 3, outputID3, ncm.InStateOutput())

	// IN Alias output
	aliasOutput1 := &iotago.AliasOutput{StateIndex: 1}
	aliasOutput2 := &iotago.AliasOutput{StateIndex: 2}
	aliasOutput3 := &iotago.AliasOutput{StateIndex: 3}

	ncm.InAliasOutput().IncMessages(aliasOutput1)
	cncm1.InAliasOutput().IncMessages(aliasOutput2)
	cncm1.InAliasOutput().IncMessages(aliasOutput3)

	checkMetricsValues(t, 2, aliasOutput3, cncm1.InAliasOutput())
	checkMetricsValues(t, 0, new(iotago.AliasOutput), cncm2.InAliasOutput())
	checkMetricsValues(t, 3, aliasOutput3, ncm.InAliasOutput())

	// IN Output
	inOutput1 := &InOutput{OutputID: iotago.OutputID{1}}
	inOutput2 := &InOutput{OutputID: iotago.OutputID{2}}
	inOutput3 := &InOutput{OutputID: iotago.OutputID{3}}

	cncm1.InOutput().IncMessages(inOutput1)
	cncm2.InOutput().IncMessages(inOutput2)
	ncm.InOutput().IncMessages(inOutput3)

	checkMetricsValues(t, 1, inOutput1, cncm1.InOutput())
	checkMetricsValues(t, 1, inOutput2, cncm2.InOutput())
	checkMetricsValues(t, 3, inOutput3, ncm.InOutput())

	// IN Transaction inclusion state
	txInclusionState1 := &TxInclusionStateMsg{TxID: iotago.TransactionID{1}}
	txInclusionState2 := &TxInclusionStateMsg{TxID: iotago.TransactionID{2}}
	txInclusionState3 := &TxInclusionStateMsg{TxID: iotago.TransactionID{3}}

	cncm1.InTxInclusionState().IncMessages(txInclusionState1)
	cncm1.InTxInclusionState().IncMessages(txInclusionState2)
	cncm2.InTxInclusionState().IncMessages(txInclusionState3)

	checkMetricsValues(t, 2, txInclusionState2, cncm1.InTxInclusionState())
	checkMetricsValues(t, 1, txInclusionState3, cncm2.InTxInclusionState())
	checkMetricsValues(t, 3, txInclusionState3, ncm.InTxInclusionState())

	// IN On ledger request

	onLedgerRequest1 := createOnLedgerRequest()
	onLedgerRequest2 := createOnLedgerRequest()
	onLedgerRequest3 := createOnLedgerRequest()

	cncm1.InOnLedgerRequest().IncMessages(onLedgerRequest1)
	cncm2.InOnLedgerRequest().IncMessages(onLedgerRequest2)
	cncm1.InOnLedgerRequest().IncMessages(onLedgerRequest3)

	checkMetricsValues(t, 2, onLedgerRequest3, cncm1.InOnLedgerRequest())
	checkMetricsValues(t, 1, onLedgerRequest2, cncm2.InOnLedgerRequest())
	checkMetricsValues(t, 3, onLedgerRequest3, ncm.InOnLedgerRequest())

	// OUT Publish state transaction
	stateTransaction1 := &StateTransaction{StateIndex: 1}
	stateTransaction2 := &StateTransaction{StateIndex: 1}
	stateTransaction3 := &StateTransaction{StateIndex: 1}

	cncm1.OutPublishStateTransaction().IncMessages(stateTransaction1)
	cncm2.OutPublishStateTransaction().IncMessages(stateTransaction2)
	cncm2.OutPublishStateTransaction().IncMessages(stateTransaction3)

	checkMetricsValues(t, 1, stateTransaction1, cncm1.OutPublishStateTransaction())
	checkMetricsValues(t, 2, stateTransaction3, cncm2.OutPublishStateTransaction())
	checkMetricsValues(t, 3, stateTransaction3, ncm.OutPublishStateTransaction())

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

	cncm2.OutPublishGovernanceTransaction().IncMessages(publishStateTransaction1)
	cncm2.OutPublishGovernanceTransaction().IncMessages(publishStateTransaction2)
	cncm1.OutPublishGovernanceTransaction().IncMessages(publishStateTransaction3)

	checkMetricsValues(t, 1, publishStateTransaction3, cncm1.OutPublishGovernanceTransaction())
	checkMetricsValues(t, 2, publishStateTransaction2, cncm2.OutPublishGovernanceTransaction())
	checkMetricsValues(t, 3, publishStateTransaction3, ncm.OutPublishGovernanceTransaction())

	// OUT Pull latest output
	ncm.OutPullLatestOutput().IncMessages("OutPullLatestOutput1")
	cncm1.OutPullLatestOutput().IncMessages("OutPullLatestOutput2")
	cncm2.OutPullLatestOutput().IncMessages("OutPullLatestOutput3")

	checkMetricsValues(t, 1, "OutPullLatestOutput2", cncm1.OutPullLatestOutput())
	checkMetricsValues(t, 1, "OutPullLatestOutput3", cncm2.OutPullLatestOutput())
	checkMetricsValues(t, 3, "OutPullLatestOutput3", ncm.OutPullLatestOutput())

	// OUT Pull transaction inclusion state
	transactionID1 := iotago.TransactionID{1}
	transactionID2 := iotago.TransactionID{2}
	transactionID3 := iotago.TransactionID{3}

	cncm1.OutPullTxInclusionState().IncMessages(transactionID1)
	ncm.OutPullTxInclusionState().IncMessages(transactionID2)
	cncm2.OutPullTxInclusionState().IncMessages(transactionID3)

	checkMetricsValues(t, 1, transactionID1, cncm1.OutPullTxInclusionState())
	checkMetricsValues(t, 1, transactionID3, cncm2.OutPullTxInclusionState())
	checkMetricsValues(t, 3, transactionID3, ncm.OutPullTxInclusionState())

	// OUT Pull output by ID
	utxoInput1 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}
	utxoInput2 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}
	utxoInput3 := &iotago.UTXOInput{TransactionID: iotago.TransactionID{1}}

	cncm1.OutPullOutputByID().IncMessages(utxoInput1.ID())
	cncm1.OutPullOutputByID().IncMessages(utxoInput2.ID())
	cncm1.OutPullOutputByID().IncMessages(utxoInput3.ID())

	checkMetricsValues(t, 3, utxoInput3.ID(), cncm1.OutPullOutputByID())
	checkMetricsValues(t, 0, iotago.OutputID{}, cncm2.OutPullOutputByID())
	checkMetricsValues(t, 3, utxoInput3.ID(), ncm.OutPullOutputByID())

	// IN Milestone
	milestoneInfo1 := &nodeclient.MilestoneInfo{Index: 0}
	milestoneInfo2 := &nodeclient.MilestoneInfo{Index: 0}

	ncm.InMilestone().IncMessages(milestoneInfo1)
	ncm.InMilestone().IncMessages(milestoneInfo2)

	checkMetricsValues(t, 2, milestoneInfo2, ncm.InMilestone())
}

func checkMetricsValues[T any, V any](t *testing.T, expectedTotal uint32, expectedLastMessage V, metrics IMessageMetric[T]) {
	require.Equal(t, expectedTotal, metrics.MessagesTotal())
	if expectedTotal == 0 {
		var zeroValue V
		require.Equal(t, zeroValue, metrics.LastMessage())
	} else {
		require.Equal(t, expectedLastMessage, metrics.LastMessage())
	}
}

func TestPeeringMetrics(t *testing.T) {
	pmp := NewPeeringMetricsProvider()

	pmp.RecvEnqueued(100, 1)
	pmp.RecvEnqueued(1009, 2)
	pmp.RecvDequeued(1009, 1)
	pmp.RecvEnqueued(100, 0)

	pmp.SendEnqueued(100, 1)
	pmp.SendEnqueued(1009, 2)
	pmp.SendDequeued(1009, 1)
	pmp.SendEnqueued(100, 0)
}
