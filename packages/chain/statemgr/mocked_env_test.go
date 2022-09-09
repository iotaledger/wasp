// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledgers           *testchain.MockedLedgers
	OriginatorKeyPair *cryptolib.KeyPair
	OriginatorAddress iotago.Address
	StateKeyPair      *cryptolib.KeyPair
	NodeIDs           []string
	NodePubKeys       []*cryptolib.PublicKey
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	ChainID           *isc.ChainID
	mutex             sync.Mutex
	Nodes             map[cryptolib.PublicKeyKey]*MockedNode
}

func NewMockedEnv(nodeCount int, t *testing.T, debug bool) *MockedEnv {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t), level, false)
	ret := &MockedEnv{
		T:                 t,
		Log:               log,
		OriginatorKeyPair: nil,
		OriginatorAddress: nil,
		Nodes:             make(map[cryptolib.PublicKeyKey]*MockedNode),
	}

	ret.StateKeyPair = cryptolib.NewKeyPair()
	ret.Ledgers = testchain.NewMockedLedgers(log)
	ret.ChainID = ret.Ledgers.InitLedger(ret.StateKeyPair.GetPublicKey().AsEd25519Address())

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	nodeIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	ret.NodeIDs = nodeIDs
	ret.NodePubKeys = testpeers.PublicKeys(nodeIdentities)
	ret.NetworkProviders, _ = testpeers.SetupNet(nodeIDs, nodeIdentities, ret.NetworkBehaviour, log)

	log.Infof("Testing environment is ready")

	return ret
}

func (env *MockedEnv) SetPushStateToNodesOption(push bool) {
	env.Ledgers.SetPushOutputToNodesNeeded(push)
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[node.PubKey.AsKey()]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.PubKey.String())
	}
	env.Nodes[node.PubKey.AsKey()] = node
}

func (env *MockedEnv) RemoveNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, node.PubKey.AsKey())
}
