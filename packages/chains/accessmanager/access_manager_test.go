// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessmanager_test

import (
	"context"
	"fmt"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hivelog "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/chains/accessmanager"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/util"
)

type tc struct {
	n        int
	reliable bool
}

func TestBasic(t *testing.T) {
	t.Parallel()
	tests := []tc{
		{n: 1, reliable: true},  // Low N
		{n: 2, reliable: true},  // Low N
		{n: 3, reliable: true},  // Low N
		{n: 4, reliable: true},  // Minimal robust config.
		{n: 10, reliable: true}, // Typical config.
	}
	if !testing.Short() {
		tests = append(tests,
			tc{n: 4, reliable: false},  // Minimal robust config.
			tc{n: 10, reliable: false}, // Typical config.
			tc{n: 31, reliable: true},  // Large cluster, reliable - to make test faster.
		)
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,Reliable=%v", tst.n, tst.reliable),
			func(tt *testing.T) { testBasic(tt, tst.n, tst.reliable) },
		)
	}
}

func testBasic(t *testing.T, n int, reliable bool) {
	t.Parallel()
	ctx, ctxCancel := context.WithCancel(context.Background())
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer ctxCancel()

	peeringURLs, peerIdentities := testpeers.SetupKeys(uint16(n))
	peerPubKeys := testpeers.PublicKeys(peerIdentities)
	var networkBehaviour testutil.PeeringNetBehavior
	if reliable {
		networkBehaviour = testutil.NewPeeringNetReliable(log)
	} else {
		netLogger := testlogger.WithLevel(log.NewChildLogger("Network"), hivelog.LevelInfo, false)
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, netLogger)
	}
	peeringNetwork := testutil.NewPeeringNetwork(
		peeringURLs, peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(log, hivelog.LevelWarning, false),
	)
	networkProviders := peeringNetwork.NetworkProviders()
	defer peeringNetwork.Close()

	accessMgrs := make([]accessmanager.AccessMgr, len(peerIdentities))
	nodeServers := make([][]*cryptolib.PublicKey, len(peerIdentities)) // That's the output.
	for i := range accessMgrs {
		ii := i
		serversUpdatedCB := func(chainID isc.ChainID, servers []*cryptolib.PublicKey) {
			t.Logf("servers updated, ChainID=%v, servers=%+v", chainID, servers)
			nodeServers[ii] = servers
		}
		accessMgrs[i] = accessmanager.New(ctx, serversUpdatedCB, peerIdentities[i], networkProviders[i], log.NewChildLogger(fmt.Sprintf("N#%v", i)))
	}
	//
	// Make all of them trusted.
	for _, am := range accessMgrs {
		am.TrustedNodes(peerPubKeys)
	}
	//
	// Everyone gives access to everyone.
	chainID := isctest.RandomChainID()
	for _, am := range accessMgrs {
		am.ChainAccessNodes(chainID, peerPubKeys)
	}

	ctx, cancel := context.WithTimeout(context.Background(), testmisc.GetTimeout(1*time.Minute))
	defer cancel()

	//
	// Wait for everyone to get the server nodes.
	for done := false; !done; {
		require.NoError(t, ctx.Err(), "timeout: wait for everyone to get the server nodes")

		time.Sleep(100 * time.Millisecond)
		done = true
		for i := range nodeServers {
			if !util.Same(nodeServers[i], peerPubKeys) {
				t.Logf("Wait for node %v", i)
				done = false
				break
			}
		}
	}
}
