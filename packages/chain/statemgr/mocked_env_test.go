// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	// "fmt"
	"io"
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	// "github.com/iotaledger/wasp/packages/transaction"
	// "github.com/iotaledger/wasp/packages/utxodb"
	// "github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledger            *testchain.MockedLedger
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress iotago.Address
	StateKeyPair      *ed25519.KeyPair
	NodePubKeys       []*ed25519.PublicKey
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	NetworkCloser     io.Closer
	ChainID           *iscp.ChainID
	mutex             sync.Mutex
	Nodes             map[ed25519.PublicKey]*MockedNode
	push              bool
}

func NewMockedEnv(nodeCount int, t *testing.T, debug bool) (*MockedEnv, *iotago.Transaction) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)
	ret := &MockedEnv{
		T:                 t,
		Log:               log,
		OriginatorKeyPair: nil,
		OriginatorAddress: nil,
		Nodes:             make(map[ed25519.PublicKey]*MockedNode),
	}
	/*origKeyPairCryptolib, origAddr := ret.Ledger.NewKeyPairByIndex(0)
	stateKeyPairCryptolib, stateAddr := ret.Ledger.NewKeyPairByIndex(1)
	origKeyPairHive := cryptolib.CryptolibKeyPairToHiveKeyPair(origKeyPairCryptolib)
	stateKeyPairHive := cryptolib.CryptolibKeyPairToHiveKeyPair(stateKeyPairCryptolib)
	ret.OriginatorKeyPair = &origKeyPairHive
	ret.OriginatorAddress = origAddr
	ret.StateKeyPair = &stateKeyPairHive
	_, err := ret.Ledger.GetFundsFromFaucet(ret.OriginatorAddress)
	require.NoError(t, err)
	require.EqualValues(t, utxodb.FundsFromFaucetAmount, ret.Ledger.GetAddressBalanceIotas(origAddr))
	require.EqualValues(t, 0, ret.Ledger.GetAddressBalanceIotas(stateAddr))

	allOutputs, ids := ret.Ledger.GetUnspentOutputs(origAddr)
	originTx, chainID, err := transaction.NewChainOriginTransaction(
		origKeyPairCryptolib,
		stateAddr,
		stateAddr,
		0,
		allOutputs,
		ids,
		ret.Ledger.RentStructure(),
	)
	fmt.Printf("XXX: %T %v\n", originTx.UnlockBlocks[0], originTx.UnlockBlocks[0])
	ret.ChainID = chainID
	require.NoError(t, err)

	err = ret.Ledger.AddToLedger(originTx)
	require.NoError(t, err)

	originTxID, err := originTx.ID()
	require.NoError(t, err)

	txBack, ok := ret.Ledger.GetTransaction(*originTxID)
	require.True(t, ok)
	txidBack, err := txBack.ID()
	require.NoError(t, err)
	require.EqualValues(t, *originTxID, *txidBack)

	t.Logf("New chain ID: %s", chainID.String())

	retOut, _, err := utxodb.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)
	require.NotNil(t, retOut)*/

	keyPair := ed25519.GenerateKeyPair()
	addr := cryptolib.Ed25519AddressFromPubKey(cryptolib.HivePublicKeyToCryptolibPublicKey(keyPair.PublicKey))

	originOutput := &iotago.AliasOutput{
		Amount: iotago.TokenSupply,
		//StateController:      stateControllerAddress,
		//GovernanceController: governanceControllerAddress,
		StateMetadata: state.OriginStateHash().Bytes(),
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: addr,
			},
		},
	}
	ret.ChainID = iscp.RandomChainID()
	ret.Ledger = testchain.NewMockedLedger(originOutput)

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	nodeIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	ret.NodePubKeys = testpeers.PublicKeys(nodeIdentities)
	ret.NetworkProviders, ret.NetworkCloser = testpeers.SetupNet(nodeIDs, nodeIdentities, ret.NetworkBehaviour, log)

	return ret, nil //originTx
}

func (env *MockedEnv) SetPushStateToNodesOption(push bool) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	env.push = push
}

func (env *MockedEnv) pushStateToNodesIfSet(tx *iotago.Transaction) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if !env.push {
		return
	}
	/*	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
		require.NoError(env.T, err)

		for _, node := range env.Nodes {
			go node.StateManager.EnqueueStateMsg(&messages.StateMsg{
				ChainOutput: stateOutput,
				Timestamp:   tx.Essence().Timestamp(),
			})
		}*/
}

func (env *MockedEnv) PostTransactionToLedger(tx *iotago.Transaction) {
	panic("TODO implement")
	/*txID, err := tx.ID()
	if err != nil {
		env.Log.Errorf("MockedEnv.PostTransactionToLedger: error geting transaction id: %v", err)
	}
	env.Log.Debugf("MockedEnv.PostTransactionToLedger: transaction %v", txID)
	_, exists := env.Ledger.GetTransaction(*txID)
	if exists {
		env.Log.Debugf("MockedEnv.PostTransactionToLedger: posted repeating originTx: %v", txID)
		return
	}
	if err = env.Ledger.AddToLedger(tx); err != nil {
		env.Log.Errorf("MockedEnv.PostTransactionToLedger: error adding transaction: %v", err)
		return
	}
	// Push transaction to nodes
	go env.pushStateToNodesIfSet(tx)

	env.Log.Infof("MockedEnv.PostTransactionToLedger: posted transaction to ledger: %s", txID)*/
}

func (env *MockedEnv) PullStateFromLedger() *messages.StateMsg {
	panic("TODO implement")
	/*env.Log.Debugf("MockedEnv.PullStateFromLedger request received")
	//outputs := env.Ledger.GetAddressOutputs(env.ChainID.AsAliasAddress())
	_, outputIDs := env.Ledger.GetUnspentOutputs(env.ChainID.AsAddress())
	require.EqualValues(env.T, 1, len(outputIDs))
	outTx, ok := env.Ledger.GetTransaction(outputIDs[0].TransactionID)
	require.True(env.T, ok)
	stateOutput, stateOutputID, err := utxodb.GetSingleChainedAliasOutput(outTx)
	require.NoError(env.T, err)
	stateOutputIDOtherType := stateOutputID.UTXOInput()
	env.Log.Debugf("MockedEnv.PullStateFromLedger chain output %s found", iscp.OID(stateOutputIDOtherType))
	return &messages.StateMsg{
		ChainOutput: iscp.NewAliasOutputWithID(stateOutput, stateOutputIDOtherType),
		//Timestamp:   outTx.Essence.Timestamp(),	% No such field in tx
	}*/
}

func (env *MockedEnv) PullConfirmedOutputFromLedger(outputID *iotago.UTXOInput) iotago.Output {
	panic("TODO implement")
	/*env.Log.Debugf("MockedEnv.PullConfirmedOutputFromLedger for output %v", iscp.OID(outputID))
	tx, foundTx := env.Ledger.GetTransaction(outputID.TransactionID)
	require.True(env.T, foundTx)
	outputs, err := tx.OutputsSet()
	require.NoError(env.T, err)
	output, ok := outputs[outputID.ID()]
	require.True(env.T, ok)
	env.Log.Debugf("MockedEnv.PullConfirmedOutputFromLedger output found")
	return output*/
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[*node.PubKey]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.PubKey.String())
	}
	env.Nodes[*node.PubKey] = node
}

func (env *MockedEnv) RemoveNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, *node.PubKey)
}
