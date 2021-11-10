// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/vm/wasmproc"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"
)

// DefaultTimeStep is a default step for the logical clock for each PostRequestSync call.
const DefaultTimeStep = 1 * time.Millisecond

// Saldo is the default amount of tokens returned by the UTXODB faucet
// which is therefore the amount returned by NewKeyPairWithFunds() and such
const (
	Saldo              = utxodb.RequestFundsAmount
	DustThresholdIotas = uint64(1)
	ChainDustThreshold = uint64(100)
	MaxRequestsInBlock = 100
	timeLayout         = "04:05.000000000"
)

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T           TestContext
	logger      *logger.Logger
	dbmanager   *dbmanager.DBManager
	utxoDB      *utxodb.UtxoDB
	seed        *ed25519.Seed
	blobCache   registry.BlobCache
	glbMutex    sync.RWMutex
	ledgerMutex sync.RWMutex
	clockMutex  sync.RWMutex
	logicalTime time.Time
	timeStep    time.Duration
	chains      map[[33]byte]*Chain
	vmRunner    vm.VMRunner
	// publisher wait group
	publisherWG      sync.WaitGroup
	publisherEnabled atomic.Bool
	processorConfig  *processors.Config
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
	ChainID *iscp.ChainID

	// OriginatorAddress is the alias for OriginatorKeyPair.Address()
	OriginatorAddress ledgerstate.Address

	// OriginatorAgentID is the OriginatorAddress represented in the form of AgentID
	OriginatorAgentID *iscp.AgentID

	// ValidatorFeeTarget is the agent ID to which all fees are accrued. By default is its equal to OriginatorAddress
	ValidatorFeeTarget *iscp.AgentID

	// State ia an interface to access virtual state of the chain: the collection of key/value pairs
	State       state.VirtualStateAccess
	GlobalSync  coreutil.ChainStateSync
	StateReader state.OptimisticStateReader

	// Log is the named logger of the chain
	Log *logger.Logger

	// processor cache
	proc *processors.Cache

	// related to asynchronous backlog processing
	runVMMutex sync.Mutex
	mempool    chain.Mempool
}

// New creates an instance of the `solo` environment.
//
// If solo is used for unit testing, 't' should be the *testing.T instance;
// otherwise it can be either nil or an instance created with NewTestContext.
//
// 'debug' parameter 'true' means logging level is 'debug', otherwise 'info'
// 'printStackTrace' controls printing stack trace in case of errors
func New(t TestContext, debug, printStackTrace bool, seedOpt ...*ed25519.Seed) *Solo {
	log := testlogger.NewNamedLogger(t.Name(), timeLayout)
	if !debug {
		log = testlogger.WithLevel(log, zapcore.InfoLevel, printStackTrace)
	}
	return NewWithLogger(t, log, seedOpt...)
}

// New creates an instance of the `solo` environment with the given logger.
//
// If solo is used for unit testing, 't' should be the *testing.T instance;
// otherwise it can be either nil or an instance created with NewTestContext.
func NewWithLogger(t TestContext, log *logger.Logger, seedOpt ...*ed25519.Seed) *Solo {
	if t == nil {
		t = NewTestContext("solo")
	}
	var seed *ed25519.Seed
	if len(seedOpt) > 0 {
		seed = seedOpt[0]
	}

	processorConfig := processors.NewConfig()
	err := processorConfig.RegisterVMType(vmtypes.WasmTime, func(binary []byte) (iscp.VMProcessor, error) {
		return wasmproc.GetProcessor(binary, log)
	})
	require.NoError(t, err)

	initialTime := time.Unix(1, 0)
	ret := &Solo{
		T:               t,
		logger:          log,
		dbmanager:       dbmanager.NewDBManager(log.Named("db"), true),
		utxoDB:          utxodb.NewWithTimestamp(initialTime),
		seed:            seed,
		blobCache:       iscp.NewInMemoryBlobCache(),
		logicalTime:     initialTime,
		timeStep:        DefaultTimeStep,
		chains:          make(map[[33]byte]*Chain),
		vmRunner:        runvm.NewVMRunner(),
		processorConfig: processorConfig,
	}
	ret.logger.Infof("Solo environment has been created with initial logical time %v", initialTime.Format(timeLayout))
	return ret
}

func (env *Solo) SyncLog() {
	_ = env.logger.Sync()
}

// WithNativeContract registers a native contract so that it may be deployed
func (env *Solo) WithNativeContract(c *coreutil.ContractProcessor) *Solo {
	env.processorConfig.RegisterNativeContract(c)
	return env
}

// NewChain deploys new chain instance.
//
// If 'chainOriginator' is nil, new one is generated and solo.Saldo (=1337) iotas are loaded from the UTXODB faucet.
// If 'validatorFeeTarget' is skipped, it is assumed equal to OriginatorAgentID
// To deploy a chain instance the following steps are performed:
//  - chain signature scheme (private key), chain address and chain ID are created
//  - empty virtual state is initialized
//  - origin transaction is created by the originator and added to the UTXODB
//  - 'init' request transaction to the 'root' contract is created and added to UTXODB
//  - backlog processing threads (goroutines) are started
//  - VM processor cache is initialized
//  - 'init' request is run by the VM. The 'root' contracts deploys the rest of the core contracts:
//    '_default', 'blocklog', 'blob', 'accounts' and 'eventlog',
// Upon return, the chain is fully functional to process requests
//nolint:funlen
func (env *Solo) NewChain(chainOriginator *ed25519.KeyPair, name string, validatorFeeTarget ...*iscp.AgentID) *Chain {
	env.logger.Debugf("deploying new chain '%s'", name)
	var stateController ed25519.KeyPair
	if env.seed == nil {
		stateController = ed25519.GenerateKeyPair() // chain address will be ED25519, not BLS
	} else {
		stateController = *env.seed.KeyPair(2)
	}
	stateAddr := ledgerstate.NewED25519Address(stateController.PublicKey)

	var originatorAddr ledgerstate.Address
	if chainOriginator == nil {
		if env.seed == nil {
			kp := ed25519.GenerateKeyPair()
			chainOriginator = &kp
			originatorAddr = ledgerstate.NewED25519Address(kp.PublicKey)
		} else {
			chainOriginator = env.seed.KeyPair(1)
			originatorAddr = ledgerstate.NewED25519Address(chainOriginator.PublicKey)
		}
		_, err := env.utxoDB.RequestFunds(originatorAddr, env.LogicalTime())
		require.NoError(env.T, err)
	} else {
		originatorAddr = ledgerstate.NewED25519Address(chainOriginator.PublicKey)
	}
	originatorAgentID := iscp.NewAgentID(originatorAddr, 0)
	feeTarget := originatorAgentID
	if len(validatorFeeTarget) > 0 {
		feeTarget = validatorFeeTarget[0]
	}

	bals := colored.NewBalancesForIotas(100)

	inputs := env.utxoDB.GetAddressOutputs(originatorAddr)
	originTx, chainID, err := transaction.NewChainOriginTransaction(chainOriginator, stateAddr, bals, env.LogicalTime(), inputs...)
	require.NoError(env.T, err)
	err = env.utxoDB.AddTransaction(originTx)
	require.NoError(env.T, err)
	env.AssertAddressBalance(originatorAddr, colored.IOTA, Saldo-100)

	env.logger.Infof("deploying new chain '%s'. ID: %s, state controller address: %s",
		name, chainID.String(), stateAddr.Base58())
	env.logger.Infof("     chain '%s'. state controller address: %s", chainID.String(), stateAddr.Base58())
	env.logger.Infof("     chain '%s'. originator address: %s", chainID.String(), originatorAddr.Base58())

	chainlog := env.logger.Named(name)
	store := env.dbmanager.GetOrCreateKVStore(chainID)
	vs, err := state.CreateOriginState(store, chainID)
	env.logger.Infof("     chain '%s'. origin state hash: %s", chainID.String(), vs.StateCommitment().String())

	require.NoError(env.T, err)
	require.EqualValues(env.T, 0, vs.BlockIndex())
	require.True(env.T, vs.Timestamp().IsZero())

	glbSync := coreutil.NewChainStateSync().SetSolidIndex(0)
	srdr := state.NewOptimisticStateReader(store, glbSync)

	ret := &Chain{
		Env:                    env,
		Name:                   name,
		ChainID:                chainID,
		StateControllerKeyPair: &stateController,
		StateControllerAddress: stateAddr,
		OriginatorKeyPair:      chainOriginator,
		OriginatorAddress:      originatorAddr,
		OriginatorAgentID:      originatorAgentID,
		ValidatorFeeTarget:     feeTarget,
		State:                  vs,
		StateReader:            srdr,
		GlobalSync:             glbSync,
		proc:                   processors.MustNew(env.processorConfig),
		Log:                    chainlog,
	}
	ret.mempool = mempool.New(ret.StateReader, env.blobCache, chainlog, metrics.DefaultChainMetrics())
	require.NoError(env.T, err)
	require.NoError(env.T, err)

	publisher.Event.Attach(events.NewClosure(func(msgType string, parts []string) {
		if !env.publisherEnabled.Load() {
			return
		}
		msg := msgType + " " + strings.Join(parts, " ")
		env.publisherWG.Add(1)
		go func() {
			chainlog.Infof("SOLO PUBLISHER (test %s):: '%s'", env.T.Name(), msg)
			env.publisherWG.Done()
		}()
	}))

	initTx, err := transaction.NewRootInitRequestTransaction(
		ret.OriginatorKeyPair,
		chainID,
		"'solo' testing chain",
		env.LogicalTime(),
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

	// put to mempool and take back to solidify
	ret.solidifyRequest(initReq[0])

	_, err = ret.runRequestsSync(initReq, "new")
	require.NoError(env.T, err)
	ret.logRequestLastBlock()

	ret.Log.Infof("chain '%s' deployed. Chain ID: %s", ret.Name, ret.ChainID.String())
	return ret
}

// AddToLedger adds (synchronously confirms) transaction to the UTXODB ledger. Return error if it is
// invalid or double spend
func (env *Solo) AddToLedger(tx *ledgerstate.Transaction) error {
	return env.utxoDB.AddTransaction(tx)
}

// RequestsForChain parses the transaction and returns all requests contained in it which have chainID as the target
func (env *Solo) RequestsForChain(tx *ledgerstate.Transaction, chainID *iscp.ChainID) ([]iscp.Request, error) {
	env.glbMutex.RLock()
	defer env.glbMutex.RUnlock()

	m := env.requestsByChain(tx)
	ret, ok := m[chainID.Array()]
	if !ok {
		return nil, xerrors.Errorf("chain %s does not exist", chainID.String())
	}
	return ret, nil
}

func (env *Solo) requestsByChain(tx *ledgerstate.Transaction) map[[33]byte][]iscp.Request {
	sender, err := utxoutil.GetSingleSender(tx)
	require.NoError(env.T, err)
	ret := make(map[[33]byte][]iscp.Request)
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
			lst = make([]iscp.Request, 0)
		}
		mintedAmounts := colored.BalancesFromL1Map(utxoutil.GetMintedAmounts(tx))
		ret[arr] = append(lst, request.OnLedgerFromOutput(o, sender, tx.Essence().Timestamp(), mintedAmounts))
	}
	return ret
}

// EnqueueRequests adds requests contained in the transaction to mempools of respective target chains
func (env *Solo) EnqueueRequests(tx *ledgerstate.Transaction) {
	env.glbMutex.RLock()
	defer env.glbMutex.RUnlock()

	requests := env.requestsByChain(tx)

	for chidArr, reqs := range requests {
		chid, err := iscp.ChainIDFromBytes(chidArr[:])
		require.NoError(env.T, err)
		ch, ok := env.chains[chidArr]
		if !ok {
			env.logger.Infof("dispatching requests. Unknown chain: %s", chid.String())
			continue
		}
		ch.runVMMutex.Lock()

		ch.mempool.ReceiveRequests(reqs...)

		ch.runVMMutex.Unlock()
	}
}

// EnablePublisher enables Solo publisher
func (env *Solo) EnablePublisher(enable bool) {
	env.publisherEnabled.Store(enable)
}

// WaitPublisher waits until all messages are published
func (env *Solo) WaitPublisher() {
	env.publisherWG.Wait()
}

func (ch *Chain) GetChainOutput() *ledgerstate.AliasOutput {
	outs := ch.Env.utxoDB.GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(ch.Env.T, 1, len(outs))

	return outs[0]
}

// collateBatch selects requests which are not time locked
// returns batch and and 'remains unprocessed' flag
func (ch *Chain) collateBatch() []iscp.Request {
	// emulating variable sized blocks
	maxBatch := MaxRequestsInBlock - rand.Intn(MaxRequestsInBlock/3)

	ret := make([]iscp.Request, 0)
	ready := ch.mempool.ReadyNow(ch.Env.LogicalTime())
	batchSize := len(ready)
	if batchSize > maxBatch {
		batchSize = maxBatch
	}
	ready = ready[:batchSize]
	for _, req := range ready {
		// using logical clock
		if onLegderRequest, ok := req.(*request.OnLedger); ok {
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

// batchLoop mimics behavior Wasp consensus
func (ch *Chain) batchLoop() {
	for {
		ch.Sync()
		time.Sleep(50 * time.Millisecond)
	}
}

// Sync runs all ready requests
func (ch *Chain) Sync() {
	for {
		if !ch.collateAndRunBatch() {
			return
		}
	}
}

func (ch *Chain) collateAndRunBatch() bool {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	batch := ch.collateBatch()
	if len(batch) > 0 {
		_, err := ch.runRequestsNolock(batch, "batchLoop")
		if err != nil {
			ch.Log.Errorf("runRequestsSync: %v", err)
		}
		return true
	}
	return false
}

// BacklogLen is a thread-safe function to return size of the current backlog
func (ch *Chain) BacklogLen() int {
	mstats := ch.mempool.Info()
	return mstats.InBufCounter - mstats.OutPoolCounter
}

// solidifies request arguments without mempool (only for solo)
func (ch *Chain) solidifyRequest(req iscp.Request) {
	ok, err := request.SolidifyArgs(req, ch.Env.blobCache)
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
}
