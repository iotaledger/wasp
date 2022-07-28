// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committee

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

func TestCommitteeBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	nodeCount := 4
	netIDs, identities := testpeers.SetupKeys(uint16(nodeCount))
	stateAddr, dksRegistries := testpeers.SetupDkgPregenerated(t, uint16((len(netIDs)*2)/3+1), identities)
	nodes, netCloser := testpeers.SetupNet(netIDs, identities, testutil.NewPeeringNetReliable(log), log)
	net0 := nodes[0]
	dks0, err := dksRegistries[0].LoadDKShare(stateAddr)
	require.NoError(t, err)

	c, _, err := New(dks0, isc.RandomChainID(), net0, log)
	require.NoError(t, err)
	require.True(t, c.Address().Equal(stateAddr))
	require.EqualValues(t, 4, c.Size())
	require.EqualValues(t, 3, c.Quorum())

	time.Sleep(100 * time.Millisecond)
	require.True(t, c.IsReady())
	c.Close()
	require.False(t, c.IsReady())
	require.NoError(t, netCloser.Close())
}
