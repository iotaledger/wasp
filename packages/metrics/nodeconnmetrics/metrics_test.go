// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconnmetrics

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	log := testlogger.NewLogger(t)
	chainID1 := iscp.RandomChainID()
	chainID2 := iscp.RandomChainID()
	chainID3 := iscp.RandomChainID()
	ncm := New(log)

	require.Equal(t, []*iscp.ChainID{}, ncm.GetRegistered())

	ncm.SetRegistered(chainID1)
	require.Equal(t, []*iscp.ChainID{chainID1}, ncm.GetRegistered())

	ncm.SetRegistered(chainID2)
	registered := ncm.GetRegistered()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID1)
	require.Contains(t, registered, chainID2)

	ncm.SetUnregistered(chainID1)
	require.Equal(t, []*iscp.ChainID{chainID2}, ncm.GetRegistered())

	ncm.SetRegistered(chainID3)
	registered = ncm.GetRegistered()
	require.Equal(t, 2, len(registered))
	require.Contains(t, registered, chainID2)
	require.Contains(t, registered, chainID3)

	ncm.SetUnregistered(chainID3)
	require.Equal(t, []*iscp.ChainID{chainID2}, ncm.GetRegistered())

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

func TestMessageMetrics(t *testing.T) {
	log := testlogger.NewLogger(t)
	ncm := New(log)
	cncm1 := ncm.NewMessagesMetrics(iscp.RandomChainID())
	cncm2 := ncm.NewMessagesMetrics(iscp.RandomChainID())
	ncm.RegisterMetrics()

	// IN State output
	cncm1.GetInStateOutput().CountLastMessage("InStateOutput1")
	cncm1.GetInStateOutput().CountLastMessage("InStateOutput2")
	cncm1.GetInStateOutput().CountLastMessage("InStateOutput3")

	checkMetricsValues(t, 3, "InStateOutput3", cncm1.GetInStateOutput())
	checkMetricsValues(t, 0, "NIL", cncm2.GetInStateOutput())
	checkMetricsValues(t, 3, "InStateOutput3", ncm.GetInStateOutput())

	// IN Alias output
	ncm.GetInAliasOutput().CountLastMessage("InAliasOutput1")
	cncm1.GetInAliasOutput().CountLastMessage("InAliasOutput2")
	cncm1.GetInAliasOutput().CountLastMessage("InAliasOutput3")

	checkMetricsValues(t, 2, "InAliasOutput3", cncm1.GetInAliasOutput())
	checkMetricsValues(t, 0, "NIL", cncm2.GetInAliasOutput())
	checkMetricsValues(t, 3, "InAliasOutput3", ncm.GetInAliasOutput())

	// IN Output
	cncm1.GetInOutput().CountLastMessage("InOutput1")
	cncm2.GetInOutput().CountLastMessage("InOutput2")
	ncm.GetInOutput().CountLastMessage("InOutput3")

	checkMetricsValues(t, 1, "InOutput1", cncm1.GetInOutput())
	checkMetricsValues(t, 1, "InOutput2", cncm2.GetInOutput())
	checkMetricsValues(t, 3, "InOutput3", ncm.GetInOutput())

	// IN Transaction inclusion state
	cncm1.GetInTxInclusionState().CountLastMessage("InTxInclusionState1")
	cncm1.GetInTxInclusionState().CountLastMessage("InTxInclusionState2")
	cncm2.GetInTxInclusionState().CountLastMessage("InTxInclusionState3")

	checkMetricsValues(t, 2, "InTxInclusionState2", cncm1.GetInTxInclusionState())
	checkMetricsValues(t, 1, "InTxInclusionState3", cncm2.GetInTxInclusionState())
	checkMetricsValues(t, 3, "InTxInclusionState3", ncm.GetInTxInclusionState())

	// IN On ledger request
	cncm1.GetInOnLedgerRequest().CountLastMessage("InOnLedgerRequest1")
	cncm2.GetInOnLedgerRequest().CountLastMessage("InOnLedgerRequest2")
	cncm1.GetInOnLedgerRequest().CountLastMessage("InOnLedgerRequest3")

	checkMetricsValues(t, 2, "InOnLedgerRequest3", cncm1.GetInOnLedgerRequest())
	checkMetricsValues(t, 1, "InOnLedgerRequest2", cncm2.GetInOnLedgerRequest())
	checkMetricsValues(t, 3, "InOnLedgerRequest3", ncm.GetInOnLedgerRequest())

	// OUT Publish transaction
	cncm1.GetOutPublishTransaction().CountLastMessage("OutPublishTransaction1")
	cncm2.GetOutPublishTransaction().CountLastMessage("OutPublishTransaction2")
	cncm2.GetOutPublishTransaction().CountLastMessage("OutPublishTransaction3")

	checkMetricsValues(t, 1, "OutPublishTransaction1", cncm1.GetOutPublishTransaction())
	checkMetricsValues(t, 2, "OutPublishTransaction3", cncm2.GetOutPublishTransaction())
	checkMetricsValues(t, 3, "OutPublishTransaction3", ncm.GetOutPublishTransaction())

	// OUT Pull latest output
	ncm.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput1")
	cncm1.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput2")
	cncm2.GetOutPullLatestOutput().CountLastMessage("OutPullLatestOutput3")

	checkMetricsValues(t, 1, "OutPullLatestOutput2", cncm1.GetOutPullLatestOutput())
	checkMetricsValues(t, 1, "OutPullLatestOutput3", cncm2.GetOutPullLatestOutput())
	checkMetricsValues(t, 3, "OutPullLatestOutput3", ncm.GetOutPullLatestOutput())

	// OUT Pull transaction inclusion state
	cncm1.GetOutPullTxInclusionState().CountLastMessage("OutPullTxInclusionState1")
	ncm.GetOutPullTxInclusionState().CountLastMessage("OutPullTxInclusionState2")
	cncm2.GetOutPullTxInclusionState().CountLastMessage("OutPullTxInclusionState3")

	checkMetricsValues(t, 1, "OutPullTxInclusionState1", cncm1.GetOutPullTxInclusionState())
	checkMetricsValues(t, 1, "OutPullTxInclusionState3", cncm2.GetOutPullTxInclusionState())
	checkMetricsValues(t, 3, "OutPullTxInclusionState3", ncm.GetOutPullTxInclusionState())

	// OUT Pull output by ID
	cncm1.GetOutPullOutputByID().CountLastMessage("OutPullOutputByID1")
	cncm1.GetOutPullOutputByID().CountLastMessage("OutPullOutputByID2")
	cncm1.GetOutPullOutputByID().CountLastMessage("OutPullOutputByID3")

	checkMetricsValues(t, 3, "OutPullOutputByID3", cncm1.GetOutPullOutputByID())
	checkMetricsValues(t, 0, "NIL", cncm2.GetOutPullOutputByID())
	checkMetricsValues(t, 3, "OutPullOutputByID3", ncm.GetOutPullOutputByID())

	// IN Milestone
	ncm.GetInMilestone().CountLastMessage("InMilestone1")
	ncm.GetInMilestone().CountLastMessage("InMilestone2")

	checkMetricsValues(t, 2, "InMilestone2", ncm.GetInMilestone())
}

func checkMetricsValues(t *testing.T, expectedTotal uint32, expectedLastMessage string, metrics NodeConnectionMessageMetrics) {
	require.Equal(t, expectedTotal, metrics.GetMessageTotal())
	if expectedTotal == 0 {
		require.Nil(t, metrics.GetLastMessage())
	} else {
		require.Equal(t, expectedLastMessage, metrics.GetLastMessage().(string))
	}
}
