package consensus1imp

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chain/committeeimpl"
	"github.com/iotaledger/wasp/packages/dbprovider"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"

	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/mock_chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestConsensus1Impl(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	chainID := coretypes.RandomChainID()
	chainCore := mock_chain.NewMockedChainCore(*chainID, log)

	dbp := dbprovider.NewInMemoryDBProvider(log)
	virtualState, err := state.CreateOriginState(dbp, nil)
	require.NoError(t, err)
	rdr, err := state.NewStateReader(dbp, nil)
	require.NoError(t, err)

	blobCache := coretypes.NewInMemoryBlobCache()
	mpool := mempool.New(rdr, blobCache, log)

	stateKeyPair := ed25519.GenerateKeyPair()
	stateAddr := ledgerstate.NewED25519Address(stateKeyPair.PublicKey)

	neighbors := []string{"localhost:4000", "localhost:4001", "localhost:4002", "localhost:4003"}
	ownNetID := "localhost:4001"
	ownPort := 4001
	peerCfg, err := peering_pkg.NewStaticPeerNetworkConfigProvider(ownNetID, ownPort, neighbors...)
	require.NoError(t, err)
	net0, err := udp.NewNetworkProvider(peerCfg, key.NewKeyPair(suite), suite, log.Named("node0"))
	require.NoError(t, err)

	mockRegistry := mock_chain.NewMockedRegistry(4, 3, 1, neighbors)
	committee, err := committeeimpl.NewCommittee(stateAddr, net0, peerCfg, mockRegistry, mockRegistry, log)
	require.NoError(t, err)

	consensus := New(chainCore, mpool, committee, mock_chain.NewMockedNodeConnection(), log)
	require.NotNil(t, consensus)
	time.Sleep(200 * time.Millisecond)
	require.True(t, consensus.IsReady())

	consensus.EventStateTransitionMsg(&chain.StateTransitionMsg{
		VariableState: virtualState,
		ChainOutput:   nil,
		Timestamp:     time.Now(),
	})
	time.Sleep(100 * time.Millisecond)
}
