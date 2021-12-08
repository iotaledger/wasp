// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"
)

// DefaultTimeStep is a default step for the logical clock for each PostRequestSync call.
const DefaultTimeStep = 1 * time.Millisecond

// Saldo is the default amount of tokens returned by the UTXODB faucet
// which is therefore the amount returned by NewPrivateKeyWithFunds() and such
const (
	Saldo              = utxodb.RequestFundsAmount
	MaxRequestsInBlock = 100
	timeLayout         = "04:05.000000000"
)

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T                TestContext
	logger           *logger.Logger
	dbmanager        *dbmanager.DBManager
	utxoDB           *utxodb.UtxoDB
	glbMutex         sync.RWMutex
	ledgerMutex      sync.RWMutex
	clockMutex       sync.RWMutex
	logicalTime      time.Time
	timeStep         time.Duration
	chains           map[iscp.ChainID]*Chain
	vmRunner         vm.VMRunner
	publisherWG      sync.WaitGroup
	publisherEnabled atomic.Bool
	processorConfig  *processors.Config
	deSeriParams     *iotago.DeSerializationParameters
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
	StateControllerKeyPair cryptolib.KeyPair
	StateControllerAddress iotago.Address

	// ChainID is the ID of the chain (in this version alias of the ChainAddress)
	ChainID *iscp.ChainID

	// OriginatorPrivateKey the key pair used to create the chain (origin transaction).
	// It is a default key pair in many of Solo calls which require private key.
	OriginatorPrivateKey cryptolib.KeyPair
	OriginatorAddress    iotago.Address
	// OriginatorAgentID is the OriginatorAddress represented in the form of AgentID
	OriginatorAgentID *iscp.AgentID

	// ValidatorFeeTarget is the agent ID to which all fees are accrued. By default, it is equal to OriginatorAgentID
	ValidatorFeeTarget *iscp.AgentID

	// State ia an interface to access virtual state of the chain: a buffered collection of key/value pairs
	State state.VirtualStateAccess
	// GlobalSync represents global atomic flag for the optimistic state reader. In Solo it has no function
	GlobalSync coreutil.ChainStateSync
	// StateReader is the read only access to the state
	StateReader state.OptimisticStateReader

	// Log is the named logger of the chain
	Log *logger.Logger

	// global processor cache
	proc *processors.Cache

	// related to asynchronous backlog processing
	runVMMutex sync.Mutex

	// mempool of the chain is used in Solo to mimic a real node
	mempool mempool.Mempool
}

// New creates an instance of the Solo environment.
//
// If solo is used for unit testing, 't' should be the *testing.T instance;
// otherwise it can be either nil or an instance created with NewTestContext.
//
// 'debug' parameter 'true' means logging level is 'debug', otherwise 'info'
// 'printStackTrace' controls printing stack trace in case of errors
func New(t TestContext, debug, printStackTrace bool, seedOpt ...cryptolib.Seed) *Solo {
	log := testlogger.NewNamedLogger(t.Name(), timeLayout)
	if !debug {
		log = testlogger.WithLevel(log, zapcore.InfoLevel, printStackTrace)
	}
	return NewWithLogger(t, log, seedOpt...)
}

// New creates an instance of the Solo environment with the given logger
// If solo is used for unit testing, 't' should be the *testing.T instance;
// otherwise it can be either nil or an instance created with NewTestContext.
func NewWithLogger(t TestContext, log *logger.Logger, seedOpt ...cryptolib.Seed) *Solo {
	if t == nil {
		t = NewTestContext("solo")
	}
	var seed cryptolib.Seed
	if len(seedOpt) > 0 {
		seed = seedOpt[0]
	}

	processorConfig := processors.NewConfig()
	// disable wasmtime vm for now
	//err := processorConfig.RegisterVMType(vmtypes.WasmTime, func(binary []byte) (iscp.VMProcessor, error) {
	//	return wasmproc.GetProcessor(binary, log)
	//})
	//require.NoError(t, err)

	initialTime := time.Unix(1, 0)
	initParams := utxodb.DefaultInitParams(seed[:]).
		WithTimestamp(utxodb.UnixSeconds(initialTime.Unix())).
		WithRentStructure(parameters.DeSerializationParameters().RentStructure)
	ret := &Solo{
		T:               t,
		logger:          log,
		dbmanager:       dbmanager.NewDBManager(log.Named("db"), true),
		utxoDB:          utxodb.New(initParams),
		logicalTime:     initialTime,
		timeStep:        DefaultTimeStep,
		chains:          make(map[iscp.ChainID]*Chain),
		vmRunner:        runvm.NewVMRunner(),
		processorConfig: processorConfig,
		deSeriParams:    parameters.DeSerializationParameters(),
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
func (env *Solo) NewChain(chainOriginator *cryptolib.KeyPair, name string, validatorFeeTarget ...*iscp.AgentID) *Chain {
	env.logger.Debugf("deploying new chain '%s'", name)

	stateController, stateAddr := env.utxoDB.NewKeyPairByIndex(2) // cryptolib.NewKeyPairFromSeed(env.seed.SubSeed(2))

	var originatorAddr iotago.Address
	var origKeyPair cryptolib.KeyPair
	if chainOriginator == nil {
		origKeyPair, originatorAddr = env.utxoDB.NewKeyPairByIndex(1) // cryptolib.NewKeyPairFromSeed(env.seed.SubSeed(1))
		chainOriginator = &origKeyPair
		_, err := env.utxoDB.RequestFunds(originatorAddr)
		require.NoError(env.T, err)
	} else {
		originatorAddr = cryptolib.Ed25519AddressFromPubKey(chainOriginator.PublicKey)
	}
	originatorAgentID := iscp.NewAgentID(originatorAddr, 0)
	feeTarget := originatorAgentID
	if len(validatorFeeTarget) > 0 {
		feeTarget = validatorFeeTarget[0]
	}

	outs, ids := env.utxoDB.GetUnspentOutputs(originatorAddr)
	originTx, chainID, err := transaction.NewChainOriginTransaction(
		*chainOriginator,
		stateAddr,
		stateAddr,
		0, // will be adjusted to min dust deposit
		outs,
		ids,
		env.deSeriParams,
	)
	require.NoError(env.T, err)

	anchor, err := transaction.GetAnchorFromTransaction(originTx)
	require.NoError(env.T, err)

	err = env.utxoDB.AddTransaction(originTx)
	require.NoError(env.T, err)
	env.AssertAddressIotas(originatorAddr, Saldo-anchor.Deposit)

	env.logger.Infof("deploying new chain '%s'. ID: %s, state controller address: %s",
		name, chainID.String(), stateAddr.Bech32(iscp.Bech32Prefix))
	env.logger.Infof("     chain '%s'. state controller address: %s", chainID.String(), stateAddr.Bech32(iscp.Bech32Prefix))
	env.logger.Infof("     chain '%s'. originator address: %s", chainID.String(), originatorAddr.Bech32(iscp.Bech32Prefix))

	chainlog := env.logger.Named(name)
	store := env.dbmanager.GetOrCreateKVStore(chainID)
	vs, err := state.CreateOriginState(store, chainID)
	env.logger.Infof("     chain '%s'. origin state commitment: %s", chainID.String(), vs.StateCommitment().String())

	require.NoError(env.T, err)
	require.EqualValues(env.T, 0, vs.BlockIndex())
	require.True(env.T, vs.Timestamp().IsZero())

	glbSync := coreutil.NewChainStateSync().SetSolidIndex(0)
	srdr := state.NewOptimisticStateReader(store, glbSync)

	ret := &Chain{
		Env:                    env,
		Name:                   name,
		ChainID:                chainID,
		StateControllerKeyPair: stateController,
		StateControllerAddress: stateAddr,
		OriginatorPrivateKey:   *chainOriginator,
		OriginatorAddress:      originatorAddr,
		OriginatorAgentID:      originatorAgentID,
		ValidatorFeeTarget:     feeTarget,
		State:                  vs,
		StateReader:            srdr,
		GlobalSync:             glbSync,
		proc:                   processors.MustNew(env.processorConfig),
		Log:                    chainlog,
	}
	ret.mempool = mempool.New(ret.StateReader, chainlog, metrics.DefaultChainMetrics())
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

	outs, ids = env.utxoDB.GetUnspentOutputs(originatorAddr)
	initTx, err := transaction.NewRootInitRequestTransaction(
		ret.OriginatorPrivateKey,
		chainID,
		"'solo' testing chain",
		outs,
		ids,
		env.deSeriParams,
	)
	require.NoError(env.T, err)
	require.NotNil(env.T, initTx)

	err = env.utxoDB.AddTransaction(initTx)
	require.NoError(env.T, err)

	env.glbMutex.Lock()
	env.chains[*chainID] = ret
	env.glbMutex.Unlock()

	go ret.batchLoop()

	initReq, err := env.RequestsForChain(initTx, chainID)
	require.NoError(env.T, err)

	_, err = ret.runRequestsSync(initReq, "new")
	require.NoError(env.T, err)
	ret.logRequestLastBlock()

	ret.Log.Infof("chain '%s' deployed. Chain ID: %s", ret.Name, ret.ChainID.String())
	return ret
}

// AddToLedger adds (synchronously confirms) transaction to the UTXODB ledger. Return error if it is
// invalid or double spend
func (env *Solo) AddToLedger(tx *iotago.Transaction) error {
	return env.utxoDB.AddTransaction(tx)
}

// RequestsForChain parses the transaction and returns all requests contained in it which have chainID as the target
func (env *Solo) RequestsForChain(tx *iotago.Transaction, chainID *iscp.ChainID) ([]iscp.RequestData, error) {
	env.glbMutex.RLock()
	defer env.glbMutex.RUnlock()

	m := env.requestsByChain(tx)
	ret, ok := m[*chainID]
	if !ok {
		return nil, xerrors.Errorf("chain %s does not exist", chainID.String())
	}
	return ret, nil
}

// requestsByChain parses the transaction and extracts those outputs which are interpreted as a request to a chain
func (env *Solo) requestsByChain(tx *iotago.Transaction) map[iscp.ChainID][]iscp.RequestData {
	ret := make(map[iscp.ChainID][]iscp.RequestData)
	txid, err := tx.ID()
	require.NoError(env.T, err)

	for i, out := range tx.Essence.Outputs {
		if out.Type() != iotago.OutputExtended {
			// only ExtendedOutputs are interpreted right now TODO nfts and other
			continue
		}
		// wrap output into the iscp.RequestData
		odata, err := iscp.OnLedgerFromUTXO(out, &iotago.UTXOInput{
			TransactionID:          *txid,
			TransactionOutputIndex: uint16(i),
		})
		require.NoError(env.T, err)

		addr := odata.TargetAddress()
		if addr.Type() != iotago.AddressAlias {
			continue
		}
		chainID := iscp.ChainIDFromAliasID(addr.(*iotago.AliasAddress).AliasID())
		lst, ok := ret[chainID]
		if !ok {
			lst = make([]iscp.RequestData, 0)
		}
		lst = append(lst, odata)
		ret[chainID] = lst
	}
	return ret
}

// EnqueueRequests adds requests contained in the transaction to mempools of respective target chains
func (env *Solo) EnqueueRequests(tx *iotago.Transaction) {
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

func (ch *Chain) GetChainOutput() *iotago.AliasOutput {
	outs := ch.Env.utxoDB.GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(ch.Env.T, 1, len(outs))

	return outs[0]
}

// collateBatch selects requests which are not time locked
// returns batch and and 'remains unprocessed' flag
func (ch *Chain) collateBatch() []iscp.RequestData {
	// emulating variable sized blocks
	maxBatch := MaxRequestsInBlock - rand.Intn(MaxRequestsInBlock/3)

	ready := ch.mempool.ReadyNow(ch.Env.LogicalTime())
	batchSize := len(ready)
	if batchSize > maxBatch {
		batchSize = maxBatch
	}
	ret := make([]iscp.RequestData, 0)
	for _, req := range ready[:batchSize] {
		if !req.IsOffLedger() {
			timeData := req.Unwrap().UTXO().Features().TimeLock()
			if timeData != nil && timeData.Time.After(ch.Env.LogicalTime()) {
				continue
			}
		}
		ret = append(ret, req)
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
