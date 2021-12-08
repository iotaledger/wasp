// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committee

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

func TestCommitteeBasic(t *testing.T) {
	suite := tcrypto.DefaultSuite()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	nodeCount := 4
	netIDs, identities := testpeers.SetupKeys(uint16(nodeCount))
	stateAddr, dksRegistries := testpeers.SetupDkgPregenerated(t, uint16((len(netIDs)*2)/3+1), identities, suite)
	nodes, netCloser := testpeers.SetupNet(netIDs, identities, testutil.NewPeeringNetReliable(log), log)
	net0 := nodes[0]

	cmtRec := &registry.CommitteeRecord{
		Address: stateAddr,
		Nodes:   netIDs,
	}
	c, _, err := New(cmtRec, nil, net0, dksRegistries[0], log)
	require.NoError(t, err)
	require.True(t, c.Address().Equals(stateAddr))
	require.EqualValues(t, 4, c.Size())
	require.EqualValues(t, 3, c.Quorum())

	time.Sleep(100 * time.Millisecond)
	require.True(t, c.IsReady())
	c.Close()
	require.False(t, c.IsReady())
	require.NoError(t, netCloser.Close())
}
