// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// 'solo' is a package to write unit tests for ISCP contracts in Go
// Running the smart contract on 'solo' does not require the Wasp node.
// The smart contract code is run synchronously on one process.
// The smart contract is running in exactly the same code of the VM wrapper,
// virtual state access and some other modules of the system.
// It does not use Wasp plugins, committees, consensus, state manager, database, peer and node communications.
// It uses in-memory DB for virtual state and UTXODB to mock the ledger.
// It deploys default chain and all builtin contracts
// It allows self-posting of requests, with or without time locks, however only supports one chain
package solo

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
	"time"
)

const DefaultTimeStep = 1 * time.Millisecond

// Glb global structure
type Glb struct {
	T           *testing.T
	logger      *logger.Logger
	utxoDB      *utxodb.UtxoDB
	glbMutex    *sync.Mutex
	logicalTime time.Time
	timeStep    time.Duration
	chains      map[coretypes.ChainID]*Chain
	doOnce      sync.Once
}

// Chain represents individual chain
type Chain struct {
	Glb                 *Glb
	Name                string
	ChainSigScheme      signaturescheme.SignatureScheme
	OriginatorSigScheme signaturescheme.SignatureScheme
	ChainID             coretypes.ChainID
	ChainAddress        address.Address
	ChainColor          balance.Color
	OriginatorAddress   address.Address
	OriginatorAgentID   coretypes.AgentID
	ValidatorFeeTarget  coretypes.AgentID
	StateTx             *sctransaction.Transaction
	State               state.VirtualState
	Proc                *processors.ProcessorCache
	Log                 *logger.Logger
	// related to asynchronous backlog processing
	runVMMutex   *sync.Mutex
	chPosted     sync.WaitGroup
	chInRequest  chan sctransaction.RequestRef
	backlog      []sctransaction.RequestRef
	backlogMutex *sync.Mutex
	batch        []*sctransaction.RequestRef
	batchMutex   *sync.Mutex
}

var doOnce = sync.Once{}

func New(t *testing.T, debug bool, printStackTrace bool) *Glb {
	doOnce.Do(func() {
		err := processors.RegisterVMType(wasmtimevm.VMType, wasmhost.GetProcessor)
		require.NoError(t, err)
	})
	ret := &Glb{
		T:           t,
		logger:      testutil.NewLogger(t),
		utxoDB:      utxodb.New(),
		glbMutex:    &sync.Mutex{},
		logicalTime: time.Now(),
		timeStep:    DefaultTimeStep,
		chains:      make(map[coretypes.ChainID]*Chain),
	}
	if !debug {
		ret.logger = testutil.WithLevel(ret.logger, zapcore.InfoLevel, printStackTrace)
	}
	return ret
}

func (glb *Glb) NewChain(chainOriginator signaturescheme.SignatureScheme, name string, validatorFeeTarget ...coretypes.AgentID) *Chain {
	chSig := signaturescheme.ED25519(ed25519.GenerateKeyPair()) // chain address will be ED25519, not BLS
	if chainOriginator == nil {
		chainOriginator = signaturescheme.ED25519(ed25519.GenerateKeyPair())
		_, err := glb.utxoDB.RequestFunds(chainOriginator.Address())
		require.NoError(glb.T, err)
	}
	chainID := coretypes.ChainID(chSig.Address())
	originatorAgentID := coretypes.NewAgentIDFromAddress(chainOriginator.Address())
	feeTarget := originatorAgentID
	if len(validatorFeeTarget) > 0 {
		feeTarget = validatorFeeTarget[0]
	}

	ret := &Chain{
		Glb:                 glb,
		Name:                name,
		ChainSigScheme:      chSig,
		OriginatorSigScheme: chainOriginator,
		ChainAddress:        chSig.Address(),
		OriginatorAddress:   chainOriginator.Address(),
		OriginatorAgentID:   originatorAgentID,
		ValidatorFeeTarget:  feeTarget,
		ChainID:             chainID,
		State:               state.NewVirtualState(mapdb.NewMapDB(), &chainID),
		Proc:                processors.MustNew(),
		Log:                 glb.logger.Named(name),
		//
		runVMMutex:   &sync.Mutex{},
		chInRequest:  make(chan sctransaction.RequestRef),
		backlog:      make([]sctransaction.RequestRef, 0),
		backlogMutex: &sync.Mutex{},
		batch:        nil,
		batchMutex:   &sync.Mutex{},
	}
	glb.CheckUtxodbBalance(ret.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount)
	var err error
	ret.StateTx, err = origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:             ret.ChainAddress,
		OriginatorSignatureScheme: ret.OriginatorSigScheme,
		AllInputs:                 glb.utxoDB.GetAddressOutputs(ret.OriginatorAddress),
	})
	require.NoError(glb.T, err)
	require.NotNil(glb.T, ret.StateTx)
	err = glb.utxoDB.AddTransaction(ret.StateTx.Transaction)
	require.NoError(glb.T, err)

	ret.ChainColor = balance.Color(ret.StateTx.ID())

	originBlock := state.MustNewOriginBlock(&ret.ChainColor)
	err = ret.State.ApplyBlock(originBlock)
	require.NoError(glb.T, err)
	err = ret.State.CommitToDb(originBlock)
	require.NoError(glb.T, err)

	initTx, err := origin.NewRootInitRequestTransaction(origin.NewRootInitRequestTransactionParams{
		ChainID:              chainID,
		Description:          "'solo' testing chain",
		OwnerSignatureScheme: ret.OriginatorSigScheme,
		AllInputs:            glb.utxoDB.GetAddressOutputs(ret.OriginatorAddress),
	})
	require.NoError(glb.T, err)
	require.NotNil(glb.T, initTx)

	err = glb.utxoDB.AddTransaction(initTx.Transaction)
	require.NoError(glb.T, err)

	glb.glbMutex.Lock()
	glb.chains[chainID] = ret
	glb.glbMutex.Unlock()

	go ret.readRequestsLoop()
	go ret.runBatchLoop()

	_, err = ret.runBatch([]sctransaction.RequestRef{{Tx: initTx, Index: 0}}, "new")
	require.NoError(glb.T, err)

	return ret
}

func (ch *Chain) readRequestsLoop() {
	for r := range ch.chInRequest {
		ch.backlogMutex.Lock()
		ch.backlog = append(ch.backlog, r)
		ch.backlogMutex.Unlock()
		ch.chPosted.Done()
	}
}

// collateBatch selects requests which are not time locked
// returns batch and and 'remains unprocessed' flag
func (ch *Chain) collateBatch() []sctransaction.RequestRef {
	ch.backlogMutex.Lock()
	defer ch.backlogMutex.Unlock()

	ret := make([]sctransaction.RequestRef, 0)
	remain := ch.backlog[:0]
	for _, ref := range ch.backlog {
		// using logical clock
		if int64(ref.RequestSection().Timelock()) <= ch.Glb.LogicalTime().Unix() {
			if ref.RequestSection().Timelock() != 0 {
				ch.Log.Infof("unlocked time-locked request %s", ref.RequestID().String())
			}
			ret = append(ret, ref)
		} else {
			remain = append(remain, ref)
		}
	}
	ch.backlog = remain
	return ret
}

func (ch *Chain) runBatchLoop() {
	for {
		batch := ch.collateBatch()
		if len(batch) > 0 {
			_, err := ch.runBatch(batch, "runBatchLoop")
			if err != nil {
				ch.Log.Errorf("runBatch: %v", err)
			}
			continue
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (ch *Chain) backlogLen() int {
	ch.chPosted.Wait()
	ch.backlogMutex.Lock()
	defer ch.backlogMutex.Unlock()
	return len(ch.backlog)
}

// NewSigSchemeWithFunds generates new ed25519 sigscheme and requests funds from the faucet
func (glb *Glb) NewSigSchemeWithFunds() signaturescheme.SignatureScheme {
	ret := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	_, err := glb.utxoDB.RequestFunds(ret.Address())
	if err != nil {
		glb.logger.Panicf("NewSigSchemeWithFunds: %v", err)
	}
	glb.CheckUtxodbBalance(ret.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
	return ret
}
