// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledger            *testchain.MockedLedger
	OriginatorKeyPair *cryptolib.KeyPair
	OriginatorAddress iotago.Address
	StateKeyPair      *cryptolib.KeyPair
	NodePubKeys       []*cryptolib.PublicKey
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	ChainID           *iscp.ChainID
	mutex             sync.Mutex
	Nodes             map[cryptolib.PublicKeyKey]*MockedNode
}

func NewMockedEnv(nodeCount int, t *testing.T, debug bool) (*MockedEnv, *iotago.AliasOutput) {
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
	addr := ret.StateKeyPair.GetPublicKey().AsEd25519Address()

	originOutput := &iotago.AliasOutput{
		Amount:        iotago.TokenSupply,
		StateMetadata: state.OriginStateHash().Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: addr},
			&iotago.GovernorAddressUnlockCondition{Address: addr},
		},
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: addr,
			},
		},
	}
	ret.ChainID = iscp.RandomChainID()
	ret.Ledger = testchain.NewMockedLedger(originOutput, log)

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	nodeIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	ret.NodePubKeys = testpeers.PublicKeys(nodeIdentities)
	ret.NetworkProviders, _ = testpeers.SetupNet(nodeIDs, nodeIdentities, ret.NetworkBehaviour, log)

	log.Infof("Testing environment is ready")

	return ret, originOutput
}

func (env *MockedEnv) SetPushStateToNodesOption(push bool) {
	env.Ledger.SetPushTransactionToNodesNeeded(push)
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[node.PubKey.AsKey()]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.PubKey.AsString())
	}
	env.Nodes[node.PubKey.AsKey()] = node
}

func (env *MockedEnv) RemoveNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, node.PubKey.AsKey())
}
