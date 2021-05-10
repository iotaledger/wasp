package consensus1imp

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.dedis.ch/kyber/v3"

	"github.com/iotaledger/wasp/packages/testutil/testpeers"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committeeimpl"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	Suite             *pairing.SuiteBn256
	T                 *testing.T
	N                 uint16
	Quorum            uint16
	Neighbors         []string
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	StateAddress      ledgerstate.Address
	PubKeys           []kyber.Point
	PrivKeys          []kyber.Scalar
	DKSRegistries     []coretypes.DKShareRegistryProvider
	DB                *dbprovider.DBProvider
	SolidState        state.VirtualState
	StateReader       state.StateReader
	StateOutput       *ledgerstate.AliasOutput
	NodeConn          *testchain.MockedNodeConn
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             []*MockedNode
	push              bool
}

type MockedNode struct {
	OwnIndex  uint16
	Env       *MockedEnv
	ChainCore *testchain.MockedChainCore
	Mempool   chain.Mempool
	Consensus *consensusImpl
	Log       *logger.Logger
}

type mockedConsensus struct{}

func NewMockedEnv(t *testing.T, n, quorum uint16, debug bool) (*MockedEnv, *ledgerstate.Transaction) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)
	var err error

	log.Infof("creating test environment with N = %d, T = %d", n, quorum)

	neighbors := make([]string, n)
	for i := range neighbors {
		neighbors[i] = fmt.Sprintf("localhost:%d", 4000+i)
	}

	ret := &MockedEnv{
		Suite:     pairing.NewSuiteBn256(),
		T:         t,
		N:         n,
		Quorum:    quorum,
		Neighbors: neighbors,
		Log:       log,
		Ledger:    utxodb.New(),
		NodeConn:  testchain.NewMockedNodeConnection(),
		Nodes:     make([]*MockedNode, n),
	}

	_, ret.PubKeys, ret.PrivKeys = testpeers.SetupKeys(n, ret.Suite)
	ret.StateAddress, ret.DKSRegistries = testpeers.SetupDkg(t, quorum, neighbors, ret.PubKeys, ret.PrivKeys, ret.Suite, log.Named("dkg"))

	ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err = ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddNewAliasMint(bals, ret.StateAddress, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddRemainderOutputIfNeeded(ret.OriginatorAddress, nil)
	require.NoError(t, err)
	originTx, err := txBuilder.BuildWithED25519(ret.OriginatorKeyPair)
	require.NoError(t, err)
	err = ret.Ledger.AddTransaction(originTx)
	require.NoError(t, err)

	ret.StateOutput, err = utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	ret.ChainID = *coretypes.NewChainID(ret.StateOutput.GetAliasAddress())

	ret.DB = dbprovider.NewInMemoryDBProvider(log)
	ret.SolidState, err = state.CreateOriginState(ret.DB, &ret.ChainID)
	require.NoError(t, err)
	ret.StateReader, err = state.NewStateReader(ret.DB, &ret.ChainID)
	require.NoError(t, err)

	ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction, _ ledgerstate.Address, _ uint16) {
		_, exists := ret.Ledger.GetTransaction(tx.ID())
		if exists {
			ret.Log.Debugf("posted repeating originTx: %s", tx.ID().Base58())
			return
		}
		if err := ret.Ledger.AddTransaction(tx); err != nil {
			ret.Log.Error(err)
			return
		}

		ret.Log.Infof("MockedEnv: posted transaction to ledger: %s", tx.ID().Base58())
	})
	pullBacklogOutputClosure := func(addr *ledgerstate.AliasAddress) {}
	ret.NodeConn.OnPullBacklog(pullBacklogOutputClosure)

	for i := range ret.Nodes {
		ret.Nodes[i] = ret.newNode(uint16(i))
	}

	return ret, originTx
}

func (env *MockedEnv) newNode(i uint16) *MockedNode {
	log := env.Log.Named(fmt.Sprintf("%d", i))
	chainCore := testchain.NewMockedChainCore(env.ChainID, log)
	mpool := mempool.New(env.StateReader, coretypes.NewInMemoryBlobCache(), log)
	mockCommitteeRegistry := testchain.NewMockedCommitteeRegistry(env.Neighbors)
	cfg, err := peering.NewStaticPeerNetworkConfigProvider(env.Neighbors[i], 4000+int(i), env.Neighbors...)
	require.NoError(env.T, err)
	keyPair := &key.Pair{
		Public:  env.PubKeys[i],
		Private: env.PrivKeys[i],
	}
	netObj, err := udp.NewNetworkProvider(cfg, keyPair, env.Suite, log)
	require.NoError(env.T, err)

	committee, err := committeeimpl.NewCommittee(env.StateAddress, netObj, cfg, env.DKSRegistries[i], mockCommitteeRegistry, log)
	require.NoError(env.T, err)

	ret := &MockedNode{
		OwnIndex:  i,
		Env:       env,
		ChainCore: chainCore,
		Mempool:   mpool,
		Consensus: New(chainCore, mpool, committee, testchain.NewMockedNodeConnection(), log),
		Log:       log,
	}
	return ret
}

func (env *MockedEnv) eventStateTransition() {
	nowis := time.Now()
	for _, node := range env.Nodes {
		go node.Consensus.EventStateTransitionMsg(&chain.StateTransitionMsg{
			State:          env.SolidState.Clone(),
			StateOutput:    env.StateOutput,
			StateTimestamp: nowis,
		})
	}
}

func (m *mockedConsensus) run() {

}
