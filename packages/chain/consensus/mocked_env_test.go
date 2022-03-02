// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	iotago "github.com/iotaledger/iota.go/v3"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	T                 *testing.T
	Quorum            uint16
	Log               *logger.Logger
	Ledger            *testchain.MockedLedger
	StateAddress      iotago.Address
	OriginatorKeyPair *cryptolib.KeyPair
	OriginatorAddress iotago.Address
	NodeIDs           []string
	NodePubKeys       []*cryptolib.PublicKey
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	NetworkCloser     io.Closer
	DKSRegistries     []registry.DKShareRegistryProvider
	ChainID           *iscp.ChainID
	MockedACS         chain.AsynchronousCommonSubsetRunner
	InitStateOutput   *iscp.AliasOutputWithID
	mutex             sync.Mutex
	Nodes             []*mockedNode
}

func NewMockedEnv(t *testing.T, n, quorum uint16, debug bool) *MockedEnv {
	return newMockedEnv(t, n, quorum, debug, false)
}

func NewMockedEnvWithMockedACS(t *testing.T, n, quorum uint16, debug bool) *MockedEnv {
	return newMockedEnv(t, n, quorum, debug, true)
}

func newMockedEnv(t *testing.T, n, quorum uint16, debug, mockACS bool) *MockedEnv {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)

	log.Infof("creating test environment with N = %d, T = %d", n, quorum)

	ret := &MockedEnv{
		T:      t,
		Quorum: quorum,
		Log:    log,
		Nodes:  make([]*mockedNode, n),
	}

	if mockACS {
		ret.MockedACS = testchain.NewMockedACSRunner(quorum, log)
		log.Infof("running MOCKED ACS consensus")
	} else {
		log.Infof("running REAL ACS consensus")
	}

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	log.Infof("running DKG and setting up mocked network..")
	nodeIDs, nodeIdentities := testpeers.SetupKeys(n)
	ret.NodeIDs = nodeIDs
	ret.NodePubKeys = make([]*cryptolib.PublicKey, len(nodeIdentities))
	for i := range nodeIdentities {
		ret.NodePubKeys[i] = nodeIdentities[i].GetPublicKey()
	}
	ret.StateAddress, ret.DKSRegistries = testpeers.SetupDkgPregenerated(t, quorum, nodeIdentities, tcrypto.DefaultSuite())
	ret.NetworkProviders, ret.NetworkCloser = testpeers.SetupNet(nodeIDs, nodeIdentities, ret.NetworkBehaviour, log)

	output := &iotago.AliasOutput{
		Amount:        iotago.TokenSupply,
		StateMetadata: state.OriginStateHash().Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: ret.StateAddress},
			&iotago.GovernorAddressUnlockCondition{Address: ret.StateAddress},
		},
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: ret.StateAddress,
			},
		},
	}
	ret.Ledger = testchain.NewMockedLedger(output, log)
	//ret.InitStateOutput = ret.Ledger.PullState()

	/*ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err = ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := colored.ToL1Map(colored.NewBalancesForIotas(100))

	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddNewAliasMint(bals, ret.StateAddress, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddRemainderOutputIfNeeded(ret.OriginatorAddress, nil)
	require.NoError(t, err)
	originTx, err := txBuilder.BuildWithED25519(ret.OriginatorKeyPair)
	require.NoError(t, err)
	err = ret.Ledger.AddTransaction(originTx)
	require.NoError(t, err)

	ret.InitStateOutput, err = utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	ret.ChainID = iscp.ChainIDFromAliasID(ret.InitStateOutput.GetAliasAddress())*/
	ret.ChainID = iscp.RandomChainID()

	return ret
}

func (env *MockedEnv) CreateNodes(timers ConsensusTimers) {
	for i := range env.Nodes {
		env.Nodes[i] = NewNode(env, uint16(i), timers)
	}
}

func (env *MockedEnv) nodeCount() int {
	return len(env.NodeIDs)
}

func (env *MockedEnv) SetInitialConsensusState() {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	for _, node := range env.Nodes {
		go func(n *mockedNode) {
			if n.SolidState != nil && n.SolidState.BlockIndex() == 0 {
				n.EventStateTransition()
			}
		}(node)
	}
}

func (env *MockedEnv) StartTimers() {
	for _, n := range env.Nodes {
		n.StartTimer()
	}
}

func (env *MockedEnv) WaitTimerTick(until int) error {
	checkTimerTickFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.TimerTick >= until {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodes("TimerTick", checkTimerTickFun)
}

func (env *MockedEnv) WaitStateIndex(quorum int, stateIndex uint32, timeout ...time.Duration) error {
	checkStateIndexFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.StateIndex >= stateIndex {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodesQuorum("stateIndex", quorum, checkStateIndexFun, timeout...)
}

func (env *MockedEnv) WaitMempool(numRequests int, quorum int, timeout ...time.Duration) error { //nolint:gocritic
	checkMempoolFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.Mempool.InPoolCounter >= numRequests && snap.Mempool.OutPoolCounter >= numRequests {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodesQuorum("mempool", quorum, checkMempoolFun, timeout...)
}

func (env *MockedEnv) WaitForEventFromNodes(waitName string, nodeConditionFun func(node *mockedNode) bool, timeout ...time.Duration) error {
	return env.WaitForEventFromNodesQuorum(waitName, env.nodeCount(), nodeConditionFun, timeout...)
}

func (env *MockedEnv) WaitForEventFromNodesQuorum(waitName string, quorum int, isEventOccuredFun func(node *mockedNode) bool, timeout ...time.Duration) error {
	to := 10 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	ch := make(chan int)
	nodeCount := env.nodeCount()
	deadline := time.Now().Add(to)
	for _, n := range env.Nodes {
		go func(node *mockedNode) {
			for time.Now().Before(deadline) {
				if isEventOccuredFun(node) {
					ch <- 1
				}
				time.Sleep(10 * time.Millisecond)
			}
			ch <- 0
		}(n)
	}
	var sum, total int
	for n := range ch {
		sum += n
		total++
		if sum >= quorum {
			return nil
		}
		if total >= nodeCount {
			return fmt.Errorf("Wait for %s: test timed out", waitName)
		}
	}
	return fmt.Errorf("WaitMempool: timeout expired %v", to)
}

func (env *MockedEnv) PostDummyRequests(n int, randomize ...bool) {
	reqs := make([]*request.OffLedger, n)
	for i := 0; i < n; i++ {
		reqs[i] = solo.NewCallParams("dummy", "dummy", "c", i).
			NewRequestOffLedger(iscp.RandomChainID(), env.OriginatorKeyPair)
	}
	rnd := len(randomize) > 0 && randomize[0]
	for _, n := range env.Nodes {
		for _, req := range reqs {
			go func(node *mockedNode, r *request.OffLedger) {
				if rnd {
					time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				}
				node.Mempool.ReceiveRequest(r)
			}(n, req)
		}
	}
}
