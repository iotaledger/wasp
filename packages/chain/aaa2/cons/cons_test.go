// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
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
	// Chain identifiers.
	aliasID := testiotago.RandAliasID()
	chainID := isc.ChainIDFromAliasID(aliasID)
	governor := cryptolib.NewKeyPair()
	//
	// Node Identities and shared key.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
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
		nodes[nid] = cons.New(chainID, nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Provide inputs.
	t.Logf("############ Provide Inputs.")
	now := time.Now()
	ao1 := randomAliasOutputWithID(aliasID, governor.Address(), committeeAddress)
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = ao1
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("After Inputs", t.Logf)
	//
	// Provide SM and MP responses on proposals (and time data).
	t.Logf("############ Provide TimeData and Proposals from SM/MP.")
	scClientKeyPair := cryptolib.NewKeyPair()
	scRequests := []isc.Request{}
	scRequestRefs := []*isc.RequestRef{}
	for i := 0; i < 3; i++ {
		scRequest := isc.NewOffLedgerRequest(
			&chainID,
			inccounter.Contract.Hname(),
			inccounter.FuncIncCounter.Hname(),
			dict.New(), 0,
		).WithGasBudget(100).Sign(scClientKeyPair)
		scRequests = append(scRequests, scRequest)
		scRequestRefs = append(scRequestRefs, isc.RequestRefFromRequest(scRequest))
	}
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.NotNil(t, out.NeedMempoolProposal)
		require.NotNil(t, out.NeedStateMgrStateProposal)
		tc.WithMessage(cons.NewMsgMempoolProposal(nid, scRequestRefs))
		tc.WithMessage(cons.NewMsgStateMgrProposalConfirmed(nid, ao1))
		tc.WithMessage(cons.NewMsgTimeData(nid, now))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM proposals", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Provide Decided Data from SM/MP.")
	kvStore := mapdb.NewMapDB()
	virtualStateAccess, err := state.CreateOriginState(kvStore, &chainID)
	require.NoError(t, err)
	// virtualStateAccess := state.NewVirtualState(kvStore) // TODO: ...
	glb := coreutil.NewChainStateSync()
	glb.SetSolidIndex(0)
	stateBaseline := glb.GetSolidIndexBaseline()
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.NotNil(t, out.NeedMempoolRequests)
		require.NotNil(t, out.NeedStateMgrDecidedState)
		tc.WithMessage(cons.NewMsgMempoolRequests(nid, scRequests))
		tc.WithMessage(cons.NewMsgStateMgrDecidedVirtualState(nid, ao1, stateBaseline, virtualStateAccess))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM data", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Run VM.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.NotNil(t, out.NeedVMResult)
		vmTask := out.NeedVMResult
		runner := runvm.NewVMRunner()
		require.NoError(t, runner.Run(vmTask))
		// vmTask.Results = []*vm.RequestResult{}
		// vmTask.ResultTransactionEssence = &iotago.TransactionEssence{
		// 	NetworkID: iotago.NetworkIDFromString("someNetworkID"),
		// 	Payload: ,
		// }
		tc.WithMessage(cons.NewMsgVMResult(nid, vmTask))
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

func randomAliasOutputWithID(aliasID iotago.AliasID, governorAddress, stateAddress iotago.Address) *isc.AliasOutputWithID {
	id := testiotago.RandUTXOInput()
	ao := &iotago.AliasOutput{
		AliasID:       aliasID,
		StateIndex:    0,                                  // TODO: ...
		StateMetadata: state.OriginL1Commitment().Bytes(), // TODO: ...
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateAddress},
			&iotago.GovernorAddressUnlockCondition{Address: governorAddress},
		},
	}
	return isc.NewAliasOutputWithID(ao, &id)
}
