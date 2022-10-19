// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/aaa2/chainMgr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/utxodb"
)

func TestBasic(t *testing.T) {
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 4, f: 1}, // TODO: Other combinations.
	}
	for i := range tests {
		tst := tests[i]
		t.Run(
			fmt.Sprintf("N=%v,F=%v", tst.n, tst.f),
			func(tt *testing.T) { testBasic(tt, tst.n, tst.f) },
		)
	}
}

func testBasic(t *testing.T, n, f int) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create ledger accounts.
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(governor.Address())
	require.NoError(t, err)
	_, err = utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Node identities and DKG.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := make([]gpa.NodeID, len(peerIdentities))
	for i, pid := range peerIdentities {
		nodeIDs[i] = pubKeyAsNodeID(pid.GetPublicKey())
	}
	cmtAddrA, dkRegs := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	cmtAddrB, dkRegs := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, dkRegs)
	require.NotNil(t, cmtAddrA)
	require.NotNil(t, cmtAddrB)
	//
	// Chain identifiers.
	tcl := testchain.NewTestChainLedger(t, utxoDB, governor, originator)
	originAO, chainID := tcl.MakeTxChainOrigin(cmtAddrA)
	// chainInitReqs := tcl.MakeTxChainInit()
	//
	// Construct the nodes.
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		cmtLogStore := testutil.NewMockedCmtLogStore()
		cm, err := chainMgr.New(nid, *chainID, cmtLogStore, dkRegs[i], pubKeyAsNodeID, log.Named(string(nid)[:6]))
		require.NoError(t, err)
		nodes[nid] = cm.AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	tc.PrintAllStatusStrings("Started", t.Logf)
	//
	// Provide initial AO.
	initAOInputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		initAOInputs[nid] = chainMgr.NewInputAliasOutputConfirmed(originAO)
	}
	tc.WithInputs(initAOInputs)
	tc.RunAll()
	tc.PrintAllStatusStrings("Initial AO received", t.Logf)
	// TODO: ...
}

func pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID(pubKey.String())
}
