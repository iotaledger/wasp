// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/wasmproc"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// DefaultTimeStep is a default step for the logical clock for each PostRequestSync call.
const DefaultTimeStep = 1 * time.Millisecond

// Saldo is the default amount of tokens returned by the UTXODB faucet
// which is therefore the amount returned by NewKeyPairWithFunds() and such
const (
	Saldo              = uint64(1337)
	DustThresholdIotas = uint64(1)
	ChainDustThreshold = uint64(100)
	MaxRequestsInBlock = 100
)

var DustThreshold = map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: DustThresholdIotas}

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T           *testing.T
	logger      *logger.Logger
	utxoDB      *utxodb.UtxoDB
	blobCache   coretypes.BlobCache
	glbMutex    *sync.RWMutex
	ledgerMutex *sync.RWMutex
	clockMutex  *sync.RWMutex
	logicalTime time.Time
	timeStep    time.Duration
	chains      map[[33]byte]*Chain
	doOnce      sync.Once
}

// Chain represents state of individual chain.
// There may be several parallel instances of the chain in the 'solo' test
type Chain struct {
	// Env is a pointer to the global structure of the 'solo' test
	Env *Solo

	// Name is the name of the chain
	Name string

	// StateControllerKeyPair signature scheme of the chain address, the one used to control funds owned by the chain.
	// In Solo it is Ed25519 signature scheme (in full Wasp environment is is a BLS address)
	StateControllerKeyPair *ed25519.KeyPair
	StateControllerAddress ledgerstate.Address

	// OriginatorKeyPair the signature scheme used to create the chain (origin transaction).
	// It is a default signature scheme in many of 'solo' calls which require private key.
	OriginatorKeyPair *ed25519.KeyPair

	// ChainID is the ID of the chain (in this version alias of the ChainAddress)
	ChainID coretypes.ChainID

	// OriginatorAddress is the alias for OriginatorKeyPair.Address()
	OriginatorAddress ledgerstate.Address

	// OriginatorAgentID is the OriginatorAddress represented in the form of AgentID
	OriginatorAgentID coretypes.AgentID

	// ValidatorFeeTarget is the agent ID to which all fees are accrued. By default is its equal to OriginatorAddress
	ValidatorFeeTarget coretypes.AgentID

	// State ia an interface to access virtual state of the chain: the collection of key/value pairs
	State state.VirtualState

	// Log is the named logger of the chain
	Log *logger.Logger

	// processor cache
	proc *processors.ProcessorCache

	// related to asynchronous backlog processing
	runVMMutex *sync.Mutex
	reqCounter atomic.Int32
	mempool    chain.Mempool
}

var (
	doOnce    = sync.Once{}
	glbLogger *logger.Logger
)

// New creates an instance of the `solo` environment for the test instances.
//   'debug' parameter 'true' means logging level is 'debug', otherwise 'info'
//   'printStackTrace' controls printing stack trace in case of errors
func New(t *testing.T, debug bool, printStackTrace bool) *Solo {
	doOnce.Do(func() {
		glbLogger = testlogger.NewLogger(t, "04:05.000")
		if !debug {
			glbLogger = testlogger.WithLevel(glbLogger, zapcore.InfoLevel, printStackTrace)
		}
		wasmtimeConstructor := func(binary []byte) (coretypes.VMProcessor, error) {
			return wasmproc.GetProcessor(binary, glbLogger)
		}
		err := processors.RegisterVMType(wasmtimevm.VMType, wasmtimeConstructor)
		require.NoError(t, err)
	})
	ret := &Solo{
		T:           t,
		logger:      glbLogger,
		utxoDB:      utxodb.New(),
		blobCache:   coretypes.NewDummyBlobCache(),
		glbMutex:    &sync.RWMutex{},
		clockMutex:  &sync.RWMutex{},
		ledgerMutex: &sync.RWMutex{},
		logicalTime: time.Now(),
		timeStep:    DefaultTimeStep,
		chains:      make(map[[33]byte]*Chain),
	}
	return ret
}

// NewChain deploys new chain instance.
//
// If 'chainOriginator' is nil, new one is generated and solo.Saldo (=1337) iotas are loaded from the UTXODB faucet.
// If 'validatorFeeTarget' is skipped, it is assumed equal to OriginatorAgentID
// To deploy the chai instance the following steps are performed:
//  - chain signature scheme (private key), chain address and chain ID are created
//  - empty virtual state is initialized
//  - origin transaction is created by the originator and added to the UTXODB
//  - 'init' request transaction to the 'root' contract is created and added to UTXODB
//  - backlog processing threads (goroutines) are started
//  - VM processor cache is initialized
//  - 'init' request is run by the VM. The 'root' contracts deploys the rest of the core contracts:
//    'blob', 'accountsc', 'chainlog'
// Upon return, the chain is fully functional to process requests
func (env *Solo) NewChain(chainOriginator *ed25519.KeyPair, name string, validatorFeeTarget ...coretypes.AgentID) *Chain {
	env.logger.Debugf("deploying new chain '%s'", name)
	stateController := ed25519.GenerateKeyPair() // chain address will be ED25519, not BLS
	stateAddr := ledgerstate.NewED25519Address(stateController.PublicKey)

	var originatorAddr ledgerstate.Address
	if chainOriginator == nil {
		kp := ed25519.GenerateKeyPair()
		chainOriginator = &kp
		originatorAddr = ledgerstate.NewED25519Address(kp.PublicKey)
		_, err := env.utxoDB.RequestFunds(originatorAddr)
		require.NoError(env.T, err)
	} else {
		originatorAddr = ledgerstate.NewED25519Address(chainOriginator.PublicKey)
	}
	originatorAgentID := coretypes.NewAgentID(originatorAddr, 0)
	feeTarget := originatorAgentID
	if len(validatorFeeTarget) > 0 {
		feeTarget = &validatorFeeTarget[0]
	}

	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	inputs := env.utxoDB.GetAddressOutputs(originatorAddr)
	originTx, chainID, err := sctransaction.NewChainOriginTransaction(chainOriginator, stateAddr, bals, inputs...)
	require.NoError(env.T, err)
	err = env.utxoDB.AddTransaction(originTx)
	require.NoError(env.T, err)
	env.AssertAddressBalance(originatorAddr, ledgerstate.ColorIOTA, Saldo-100)

	env.logger.Infof("deploying new chain '%s'. ID: %s, state controller address: %s",
		name, chainID.String(), stateAddr.Base58())
	env.logger.Infof("     chain '%s'. state controller address: %s", chainID.String(), stateAddr.Base58())
	env.logger.Infof("     chain '%s'. originator address: %s", chainID.String(), originatorAddr.Base58())

	ret := &Chain{
		Env:                    env,
		Name:                   name,
		ChainID:                chainID,
		StateControllerKeyPair: &stateController,
		StateControllerAddress: stateAddr,
		OriginatorKeyPair:      chainOriginator,
		OriginatorAddress:      originatorAddr,
		OriginatorAgentID:      *originatorAgentID,
		ValidatorFeeTarget:     *feeTarget,
		State:                  state.NewVirtualState(mapdb.NewMapDB(), &chainID),
		proc:                   processors.MustNew(),
		Log:                    env.logger.Named(name),
		//
		runVMMutex: &sync.Mutex{},
		mempool:    mempool.New(env.blobCache),
	}
	require.NoError(env.T, err)
	require.NoError(env.T, err)

	originBlock := state.MustNewOriginBlock(originTx.ID())
	err = ret.State.ApplyBlock(originBlock)
	require.NoError(env.T, err)
	err = ret.State.CommitToDb(originBlock)
	require.NoError(env.T, err)

	initTx, err := sctransaction.NewRootInitRequestTransaction(
		ret.OriginatorKeyPair,
		chainID,
		"'solo' testing chain",
		env.utxoDB.GetAddressOutputs(ret.OriginatorAddress)...,
	)
	require.NoError(env.T, err)
	require.NotNil(env.T, initTx)

	err = env.utxoDB.AddTransaction(initTx)
	require.NoError(env.T, err)

	env.glbMutex.Lock()
	env.chains[chainID.Array()] = ret
	env.glbMutex.Unlock()

	go ret.batchLoop()

	initReq, err := env.RequestsForChain(initTx, chainID)
	require.NoError(env.T, err)

	ret.reqCounter.Add(1)
	_, err = ret.runBatch(initReq, "new")
	require.NoError(env.T, err)

	ret.Log.Infof("chain '%s' deployed. Chain ID: %s", ret.Name, ret.ChainID.String())
	return ret
}

// AddToLedger adds (synchronously confirms) transaction to the UTXODB ledger. Return error if it is
// invalid or double spend
func (env *Solo) AddToLedger(tx *ledgerstate.Transaction) error {
	return env.utxoDB.AddTransaction(tx)
}

func (env *Solo) RequestsForChain(tx *ledgerstate.Transaction, chid coretypes.ChainID) ([]coretypes.Request, error) {
	env.glbMutex.RLock()
	defer env.glbMutex.RUnlock()

	m := env.requestsByChain(tx)
	ret, ok := m[chid.Array()]
	if !ok {
		return nil, xerrors.Errorf("chain %s does not exist", chid.String())
	}
	return ret, nil
}

func (env *Solo) requestsByChain(tx *ledgerstate.Transaction) map[[33]byte][]coretypes.Request {
	sender, err := utxoutil.GetSingleSender(tx)
	require.NoError(env.T, err)
	ret := make(map[[33]byte][]coretypes.Request)
	for _, out := range tx.Essence().Outputs() {
		o, ok := out.(*ledgerstate.ExtendedLockedOutput)
		if !ok {
			continue
		}
		arr := o.Address().Array()
		if _, ok = env.chains[arr]; !ok {
			// not a chain
			continue
		}
		lst, ok := ret[arr]
		if !ok {
			lst = make([]coretypes.Request, 0)
		}
		ret[arr] = append(lst, sctransaction.RequestOnLedgerFromOutput(o, sender, utxoutil.GetMintedAmounts(tx)))
	}
	return ret
}

// EnqueueRequests dispatches requests contained in the transaction among chains
func (env *Solo) EnqueueRequests(tx *ledgerstate.Transaction) {
	env.glbMutex.RLock()
	defer env.glbMutex.RUnlock()

	requests := env.requestsByChain(tx)

	for chidArr, reqs := range requests {
		chid, err := coretypes.ChainIDFromBytes(chidArr[:])
		require.NoError(env.T, err)
		chain, ok := env.chains[chidArr]
		if !ok {
			env.logger.Infof("dispatching requests. Unknown chain: %s", chid.String())
			continue
		}
		chain.reqCounter.Add(int32(len(reqs)))
		for _, reqRef := range reqs {
			chain.addToBacklog(reqRef)
		}
	}
}

func (ch *Chain) GetChainOutput() *ledgerstate.AliasOutput {
	outs := ch.Env.utxoDB.GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(ch.Env.T, 1, len(outs))

	return outs[0]
}

func (ch *Chain) addToBacklog(r coretypes.Request) {
	ch.mempool.ReceiveRequest(r)
	if onLedgerRequest, ok := r.(*sctransaction.RequestOnLedger); ok {
		tl := onLedgerRequest.TimeLock()
		if tl.UnixNano() == 0 {
			ch.Log.Infof("added to backlog: %s", r.ID())
		} else {
			ch.Log.Infof("added to backlog: %s. Time locked for: %v", r.ID(), tl.Sub(ch.Env.LogicalTime()))
		}
	}
}

// collateBatch selects requests which are not time locked
// returns batch and and 'remains unprocessed' flag
func (ch *Chain) collateBatch() []coretypes.Request {
	// emulating variable sized blocks
	maxBatch := MaxRequestsInBlock - rand.Intn(MaxRequestsInBlock/3)

	ret := make([]coretypes.Request, 0)
	ready := ch.mempool.GetReadyList(0)
	batchSize := len(ready)
	if batchSize > maxBatch {
		batchSize = maxBatch
	}
	ready = ready[:batchSize]
	for _, req := range ready {
		// using logical clock
		if onLegderRequest, ok := req.(*sctransaction.RequestOnLedger); ok {
			if onLegderRequest.TimeLock().Before(ch.Env.LogicalTime()) {
				if !onLegderRequest.TimeLock().IsZero() {
					ch.Log.Infof("unlocked time-locked request %s", req.ID())
				}
				ret = append(ret, req)
			}
		} else {
			ret = append(ret, req)
		}
	}
	return ret
}

// batchLoop mimics leader's behavior in the Wasp committee
func (ch *Chain) batchLoop() {
	for {
		batch := ch.collateBatch()
		if len(batch) > 0 {
			_, err := ch.runBatch(batch, "batchLoop")
			if err != nil {
				ch.Log.Errorf("runBatch: %v", err)
			}
			continue
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// backlogLen is a thread-safe function to return size of the current backlog
func (ch *Chain) backlogLen() int {
	return int(ch.reqCounter.Load())
}
