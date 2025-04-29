// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/metrics"
)

func TestRegister(t *testing.T) {
	chainID1 := isctest.RandomChainID()
	chainID2 := isctest.RandomChainID()
	chainID3 := isctest.RandomChainID()
	ncm := metrics.NewChainMetricsProvider()

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
	sender := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("sender")))

	requestRef := iotatest.RandomObjectRef()
	const tokensForGas = 1 * isc.Million

	request := &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *requestRef,
		Object: &iscmove.Request{
			ID:     *requestRef.ObjectID,
			Sender: sender.Address(),
			Message: iscmove.Message{
				Contract: uint32(isc.Hn("target_contract")),
				Function: uint32(isc.Hn("entrypoint")),
			},
			AssetsBag: iscmove.AssetsBagWithBalances{
				AssetsBag: iscmove.AssetsBag{
					ID:   *iotatest.RandomAddress(),
					Size: 1,
				},
				Assets: *iscmove.NewAssets(iotajsonrpc.CoinValue(tokensForGas)),
			},
			AllowanceBCS: bcs.MustMarshal(iscmove.NewAssets(1)),
			GasBudget:    1000,
		},
	}

	onLedgerRequest1, _ := isc.OnLedgerFromMoveRequest(request, cryptolib.NewRandomAddress())
	return onLedgerRequest1
}

func TestMessageMetrics(t *testing.T) {
	p := metrics.NewChainMetricsProvider()
	ncm := p.Message
	cncm1 := p.GetChainMetrics(isctest.RandomChainID()).Message
	cncm2 := p.GetChainMetrics(isctest.RandomChainID()).Message

	// IN Anchor
	anchor1 := &metrics.StateAnchor{StateIndex: 0, StateMetadata: "", Ref: iotago.ObjectRef{}}
	anchor2 := &metrics.StateAnchor{StateIndex: 1, StateMetadata: "", Ref: iotago.ObjectRef{}}
	anchor3 := &metrics.StateAnchor{StateIndex: 2, StateMetadata: "", Ref: iotago.ObjectRef{}}

	ncm.InAnchor().IncMessages(anchor1)
	cncm1.InAnchor().IncMessages(anchor2)
	cncm1.InAnchor().IncMessages(anchor3)

	checkMetricsValues(t, 2, anchor3, cncm1.InAnchor())
	checkMetricsValues(t, 0, new(metrics.StateAnchor), cncm2.InAnchor())
	checkMetricsValues(t, 3, anchor3, ncm.InAnchor())

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
	stateTransaction1 := &metrics.StateTransaction{StateIndex: 1}
	stateTransaction2 := &metrics.StateTransaction{StateIndex: 1}
	stateTransaction3 := &metrics.StateTransaction{StateIndex: 1}

	cncm1.OutPublishStateTransaction().IncMessages(stateTransaction1)
	cncm2.OutPublishStateTransaction().IncMessages(stateTransaction2)
	cncm2.OutPublishStateTransaction().IncMessages(stateTransaction3)

	checkMetricsValues(t, 1, stateTransaction1, cncm1.OutPublishStateTransaction())
	checkMetricsValues(t, 2, stateTransaction3, cncm2.OutPublishStateTransaction())
	checkMetricsValues(t, 3, stateTransaction3, ncm.OutPublishStateTransaction())
}

func checkMetricsValues[T any, V any](t *testing.T, expectedTotal uint32, expectedLastMessage V, metrics metrics.IMessageMetric[T]) {
	require.Equal(t, expectedTotal, metrics.MessagesTotal())
	if expectedTotal == 0 {
		var zeroValue V
		require.Equal(t, zeroValue, metrics.LastMessage())
	} else {
		require.Equal(t, expectedLastMessage, metrics.LastMessage())
	}
}

func TestPeeringMetrics(t *testing.T) {
	pmp := metrics.NewPeeringMetricsProvider()

	pmp.RecvEnqueued(100, 1)
	pmp.RecvEnqueued(1009, 2)
	pmp.RecvDequeued(1009, 1)
	pmp.RecvEnqueued(100, 0)

	pmp.SendEnqueued(100, 1)
	pmp.SendEnqueued(1009, 2)
	pmp.SendDequeued(1009, 1)
	pmp.SendEnqueued(100, 0)
}
