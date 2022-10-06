// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	t.Run("N=4,F=1", func(tt *testing.T) { testBasic(tt, 4, 1) })
}

func testBasic(t *testing.T, n, f int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Node Identities and shared key.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Construct the chain on L1.
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Construct the chain on L1: Create the accounts.
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the origin TX.
	outs, outIDs := utxoDB.GetUnspentOutputs(originator.Address())
	originTX, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		committeeAddress,
		governor.Address(),
		1_000_000,
		outs,
		outIDs,
	)
	require.NoError(t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(t, err)
	require.NotNil(t, stateAnchor)
	require.NotNil(t, aliasOutput)
	ao0 := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID.UTXOInput())
	err = utxoDB.AddToLedger(originTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the Init Request TX.
	outs, outIDs = utxoDB.GetUnspentOutputs(originator.Address())
	initTX, err := transaction.NewRootInitRequestTransaction(
		originator,
		chainID,
		"my test chain",
		outs,
		outIDs,
	)
	require.NoError(t, err)
	require.NotNil(t, initTX)
	err = utxoDB.AddToLedger(initTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Find the requests (the init request).
	initReqs := []isc.Request{}
	initReqRefs := []*isc.RequestRef{}
	outs, _ = utxoDB.GetUnspentOutputs(chainID.AsAddress())
	for outID, out := range outs {
		if out.FeatureSet().MetadataFeature() == nil {
			continue // TODO: Better way to filter non-requests.
		}
		req, err := isc.OnLedgerFromUTXO(out, outID.UTXOInput())
		if err != nil {
			continue
		}
		initReqs = append(initReqs, req)
		initReqRefs = append(initReqRefs, isc.RequestRefFromRequest(req))
	}
	//
	// Construct the nodes.
	consInstID := []byte{1, 2, 3} // ID of the consensus.
	procConfig := coreprocessors.Config().WithNativeContracts(inccounter.Processor)
	procCache := processors.MustNew(procConfig)
	nodeIDs := nodeIDsFromPubKeys(testpeers.PublicKeys(peerIdentities))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.Named(string(nid))
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		nodes[nid] = cons.New(*chainID, nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Provide inputs.
	t.Logf("############ Provide Inputs.")
	now := time.Now()
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = ao0
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("After Inputs", t.Logf)
	//
	// Provide SM and MP responses on proposals (and time data).
	t.Logf("############ Provide TimeData and Proposals from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.NotNil(t, out.NeedMempoolProposal)
		require.NotNil(t, out.NeedStateMgrStateProposal)
		tc.WithMessage(cons.NewMsgMempoolProposal(nid, initReqRefs))
		tc.WithMessage(cons.NewMsgStateMgrProposalConfirmed(nid, ao0))
		tc.WithMessage(cons.NewMsgTimeData(nid, now))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM proposals", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Provide Decided Data from SM/MP.")
	kvStore := mapdb.NewMapDB()
	virtualStateAccess, err := state.CreateOriginState(kvStore, chainID)
	require.NoError(t, err)
	chainStateSync := coreutil.NewChainStateSync()
	chainStateSync.SetSolidIndex(0)
	stateBaseline := chainStateSync.GetSolidIndexBaseline()
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.NotNil(t, out.NeedMempoolRequests)
		require.NotNil(t, out.NeedStateMgrDecidedState)
		tc.WithMessage(cons.NewMsgMempoolRequests(nid, initReqs))
		tc.WithMessage(cons.NewMsgStateMgrDecidedVirtualState(nid, ao0, stateBaseline, virtualStateAccess))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM data", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Run VM, validate the result.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.NotNil(t, out.NeedVMResult)
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		require.NoError(t, runvm.NewVMRunner().Run(out.NeedVMResult))
		tc.WithMessage(cons.NewMsgVMResult(nid, out.NeedVMResult))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("All done.", t.Logf)
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Completed, out.State)
		require.True(t, out.Terminated)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.Nil(t, out.NeedVMResult)
		require.NotNil(t, out.ResultTransaction)
		require.NotNil(t, out.ResultNextAliasOutput)
		require.NotNil(t, out.ResultState)
		block, err := out.ResultState.ExtractBlock()
		require.NoError(t, err)
		require.NotNil(t, block)
		if nid == nodeIDs[0] { // Just do this once.
			require.NoError(t, utxoDB.AddToLedger(out.ResultTransaction))
		}
	}
}

func nodeIDsFromPubKeys(pubKeys []*cryptolib.PublicKey) []gpa.NodeID {
	ret := make([]gpa.NodeID, len(pubKeys))
	for i := range pubKeys {
		ret[i] = nodeIDFromPubKey(pubKeys[i])
	}
	return ret
}

func nodeIDFromPubKey(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID("N#" + pubKey.String()[:6])
}

// scClientKeyPair := cryptolib.NewKeyPair()
// scRequests := []isc.Request{}
// scRequestRefs := []*isc.RequestRef{}
// for i := 0; i < 3; i++ {
// 	scRequest := isc.NewOffLedgerRequest(
// 		chainID,
// 		inccounter.Contract.Hname(),
// 		inccounter.FuncIncCounter.Hname(),
// 		dict.New(), 0,
// 	).WithGasBudget(100).Sign(scClientKeyPair)
// 	scRequests = append(scRequests, scRequest)
// 	scRequestRefs = append(scRequestRefs, isc.RequestRefFromRequest(scRequest))
// }
