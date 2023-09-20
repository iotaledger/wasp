// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/evm/evmlogger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
)

const (
	MaxRequestsInBlock = 100
	timeLayout         = "04:05.000000000"
)

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T                               Context
	logger                          *logger.Logger
	chainStateDatabaseManager       *database.ChainStateDatabaseManager
	utxoDB                          *utxodb.UtxoDB
	chainsMutex                     sync.RWMutex
	ledgerMutex                     sync.RWMutex
	chains                          map[isc.ChainID]*Chain
	processorConfig                 *processors.Config
	disableAutoAdjustStorageDeposit bool
	enableGasBurnLogging            bool
	seed                            cryptolib.Seed
	publisher                       *publisher.Publisher
	ctx                             context.Context
}

// data to be persisted in the snapshot
type chainData struct {
	// Name is the name of the chain
	Name string

	// StateControllerKeyPair signature scheme of the chain address, the one used to control funds owned by the chain.
	// In Solo it is Ed25519 signature scheme (in full Wasp environment is is a BLS address)
	StateControllerKeyPair *cryptolib.KeyPair

	// ChainID is the ID of the chain (in this version alias of the ChainAddress)
	ChainID isc.ChainID

	// OriginatorPrivateKey the key pair used to create the chain (origin transaction).
	// It is a default key pair in many of Solo calls which require private key.
	OriginatorPrivateKey *cryptolib.KeyPair

	// ValidatorFeeTarget is the agent ID to which all fees are accrued. By default, it is equal to OriginatorAgentID
	ValidatorFeeTarget isc.AgentID

	db         kvstore.KVStore
	writeMutex *sync.Mutex
}

// Chain represents state of individual chain.
// There may be several parallel instances of the chain in the 'solo' test
type Chain struct {
	chainData

	StateControllerAddress iotago.Address
	OriginatorAddress      iotago.Address
	OriginatorAgentID      isc.AgentID

	// Env is a pointer to the global structure of the 'solo' test
	Env *Solo

	// Store is where the chain data (blocks, state) is stored
	store indexedstore.IndexedStore
	// Log is the named logger of the chain
	log *logger.Logger
	// global processor cache
	proc *processors.Cache
	// related to asynchronous backlog processing
	runVMMutex sync.Mutex
	// mempool of the chain is used in Solo to mimic a real node
	mempool Mempool

	RequestsBlock uint32

	metrics *metrics.ChainMetrics

	migrationScheme *migrations.MigrationScheme
}

var _ chain.ChainCore = &Chain{}

type InitOptions struct {
	AutoAdjustStorageDeposit bool
	Debug                    bool
	PrintStackTrace          bool
	GasBurnLogEnabled        bool
	Seed                     cryptolib.Seed
	Log                      *logger.Logger
}

func DefaultInitOptions() *InitOptions {
	return &InitOptions{
		Debug:                    false,
		PrintStackTrace:          false,
		Seed:                     cryptolib.Seed{},
		AutoAdjustStorageDeposit: false, // is OFF by default
		GasBurnLogEnabled:        true,  // is ON by default
	}
}

// New creates an instance of the Solo environment
// If solo is used for unit testing, 't' should be the *testing.T instance;
// otherwise it can be either nil or an instance created with NewTestContext.
func New(t Context, initOptions ...*InitOptions) *Solo {
	opt := DefaultInitOptions()
	if len(initOptions) > 0 {
		opt = initOptions[0]
	}
	if opt.Log == nil {
		opt.Log = testlogger.NewNamedLogger(t.Name(), timeLayout)
		if !opt.Debug {
			opt.Log = testlogger.WithLevel(opt.Log, zapcore.InfoLevel, opt.PrintStackTrace)
		}
	}
	evmlogger.Init(opt.Log)

	chainRecordRegistryProvider, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainStateDatabaseManager, err := database.NewChainStateDatabaseManager(chainRecordRegistryProvider, database.WithEngine(hivedb.EngineMapDB))
	if err != nil {
		panic(err)
	}

	utxoDBinitParams := utxodb.DefaultInitParams()
	ctx, cancelCtx := context.WithCancel(context.Background())
	t.Cleanup(cancelCtx)
	ret := &Solo{
		T:                               t,
		logger:                          opt.Log,
		chainStateDatabaseManager:       chainStateDatabaseManager,
		utxoDB:                          utxodb.New(utxoDBinitParams),
		chains:                          make(map[isc.ChainID]*Chain),
		processorConfig:                 coreprocessors.NewConfigWithCoreContracts(),
		disableAutoAdjustStorageDeposit: !opt.AutoAdjustStorageDeposit,
		enableGasBurnLogging:            opt.GasBurnLogEnabled,
		seed:                            opt.Seed,
		publisher:                       publisher.New(opt.Log.Named("publisher")),
		ctx:                             ctx,
	}
	globalTime := ret.utxoDB.GlobalTime()
	ret.logger.Infof("Solo environment has been created: logical time: %v, time step: %v",
		globalTime.Format(timeLayout), ret.utxoDB.TimeStep())

	err = ret.processorConfig.RegisterVMType(vmtypes.WasmTime, func(binaryCode []byte) (isc.VMProcessor, error) {
		return wasmhost.GetProcessor(binaryCode, opt.Log)
	})
	require.NoError(t, err)

	_ = ret.publisher.Events.Published.Hook(func(ev *publisher.ISCEvent[any]) {
		ret.logger.Infof("solo publisher: %s %s %v", ev.Kind, ev.ChainID, ev.String())
	})

	go ret.publisher.Run(ctx)

	go ret.batchLoop()

	return ret
}

func (env *Solo) batchLoop() {
	for {
		time.Sleep(50 * time.Millisecond)
		chains := func() []*Chain {
			env.chainsMutex.Lock()
			defer env.chainsMutex.Unlock()
			return lo.Values(env.chains)
		}()
		for _, ch := range chains {
			ch.collateAndRunBatch()
		}
	}
}

func (env *Solo) GetDBHash() hashing.HashValue {
	return env.chainStateDatabaseManager.DBHash()
}

func (env *Solo) SyncLog() {
	_ = env.logger.Sync()
}

func (env *Solo) Publisher() *publisher.Publisher {
	return env.publisher
}

func (env *Solo) GetChainByName(name string) *Chain {
	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()
	for _, ch := range env.chains {
		if ch.Name == name {
			return ch
		}
	}
	panic("chain not found")
}

// WithNativeContract registers a native contract so that it may be deployed
func (env *Solo) WithNativeContract(c *coreutil.ContractProcessor) *Solo {
	env.processorConfig.RegisterNativeContract(c)
	return env
}

// NewChain deploys new default chain instance.
func (env *Solo) NewChain(depositFundsForOriginator ...bool) *Chain {
	ret, _ := env.NewChainExt(nil, 0, "chain1")
	if len(depositFundsForOriginator) == 0 || depositFundsForOriginator[0] {
		// deposit some tokens for the chain originator
		err := ret.DepositAssetsToL2(isc.NewAssetsBaseTokens(5*isc.Million), nil)
		require.NoError(env.T, err)
	}
	return ret
}

func (env *Solo) deployChain(
	chainOriginator *cryptolib.KeyPair,
	initBaseTokens uint64,
	name string,
	originParams ...dict.Dict,
) (chainData, *iotago.Transaction) {
	env.logger.Debugf("deploying new chain '%s'", name)

	if chainOriginator == nil {
		chainOriginator = env.NewKeyPairFromIndex(-1000 + len(env.chains)) // making new originator for each new chain
		originatorAddr := chainOriginator.GetPublicKey().AsEd25519Address()
		_, err := env.utxoDB.GetFundsFromFaucet(originatorAddr)
		require.NoError(env.T, err)
	}

	initParams := dict.Dict{
		origin.ParamChainOwner: isc.NewAgentID(chainOriginator.Address()).Bytes(),
		// FIXME this will cause import cycle
		// origin.ParamWaspVersion: codec.EncodeString(app.Version),
	}
	if len(originParams) > 0 {
		for k, v := range originParams[0] {
			initParams[k] = v
		}
	}

	stateControllerKey := env.NewKeyPairFromIndex(-1) // leaving positive indices to user
	stateControllerAddr := stateControllerKey.GetPublicKey().AsEd25519Address()

	originatorAddr := chainOriginator.GetPublicKey().AsEd25519Address()
	originatorAgentID := isc.NewAgentID(originatorAddr)

	initialL1Balance := env.L1BaseTokens(originatorAddr)

	outs, outIDs := env.utxoDB.GetUnspentOutputs(originatorAddr)
	originTx, originAO, chainID, err := origin.NewChainOriginTransaction(
		chainOriginator,
		stateControllerAddr,
		stateControllerAddr,
		initBaseTokens, // will be adjusted to min storage deposit + DefaultMinBaseTokensOnCommonAccount
		initParams,
		outs,
		outIDs,
		0,
	)
	require.NoError(env.T, err)

	anchor, _, err := transaction.GetAnchorFromTransaction(originTx)
	require.NoError(env.T, err)

	err = env.utxoDB.AddToLedger(originTx)
	require.NoError(env.T, err)
	env.AssertL1BaseTokens(originatorAddr, initialL1Balance-anchor.Deposit)

	env.logger.Infof("deploying new chain '%s'. ID: %s, state controller address: %s",
		name, chainID.String(), stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
	env.logger.Infof("     chain '%s'. state controller address: %s", chainID.String(), stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
	env.logger.Infof("     chain '%s'. originator address: %s", chainID.String(), originatorAddr.Bech32(parameters.L1().Protocol.Bech32HRP))

	db, writeMutex, err := env.chainStateDatabaseManager.ChainStateKVStore(chainID)
	require.NoError(env.T, err)
	originAOMinSD := parameters.L1().Protocol.RentStructure.MinRent(originAO)
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(db))
	origin.InitChain(store, initParams, originAO.Amount-originAOMinSD)

	{
		block, err2 := store.LatestBlock()
		require.NoError(env.T, err2)
		env.logger.Infof("     chain '%s'. origin trie root: %s", chainID, block.TrieRoot())
	}

	return chainData{
		Name:                   name,
		ChainID:                chainID,
		StateControllerKeyPair: stateControllerKey,
		OriginatorPrivateKey:   chainOriginator,
		ValidatorFeeTarget:     originatorAgentID,
		db:                     db,
		writeMutex:             writeMutex,
	}, originTx
}

// NewChainExt returns also origin and init transactions. Used for core testing
//
// If 'chainOriginator' is nil, new one is generated and utxodb.FundsFromFaucetAmount (many) base tokens are loaded from the UTXODB faucet.
// ValidatorFeeTarget will be set to OriginatorAgentID, and can be changed after initialization.
// To deploy a chain instance the following steps are performed:
//   - chain signature scheme (private key), chain address and chain ID are created
//   - empty virtual state is initialized
//   - origin transaction is created by the originator and added to the UTXODB
//   - 'init' request transaction to the 'root' contract is created and added to UTXODB
//   - backlog processing threads (goroutines) are started
//   - VM processor cache is initialized
//   - 'init' request is run by the VM. The 'root' contracts deploys the rest of the core contracts:
//
// Upon return, the chain is fully functional to process requests
func (env *Solo) NewChainExt(
	chainOriginator *cryptolib.KeyPair,
	initBaseTokens uint64,
	name string,
	originParams ...dict.Dict,
) (*Chain, *iotago.Transaction) {
	chData, originTx := env.deployChain(chainOriginator, initBaseTokens, name, originParams...)

	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()
	ch := env.addChain(chData)

	ch.log.Infof("chain '%s' deployed. Chain ID: %s", ch.Name, ch.ChainID.String())
	return ch, originTx
}

func (env *Solo) addChain(chData chainData) *Chain {
	ch := &Chain{
		chainData:              chData,
		StateControllerAddress: chData.StateControllerKeyPair.GetPublicKey().AsEd25519Address(),
		OriginatorAddress:      chData.OriginatorPrivateKey.GetPublicKey().AsEd25519Address(),
		OriginatorAgentID:      isc.NewAgentID(chData.OriginatorPrivateKey.GetPublicKey().AsEd25519Address()),
		Env:                    env,
		store:                  indexedstore.New(state.NewStore(chData.db, chData.writeMutex)),
		proc:                   processors.MustNew(env.processorConfig),
		log:                    env.logger.Named(chData.Name),
		metrics:                metrics.NewChainMetricsProvider().GetChainMetrics(chData.ChainID),
		mempool:                newMempool(env.utxoDB.GlobalTime, chData.ChainID),
		migrationScheme:        allmigrations.DefaultScheme,
	}
	env.chains[chData.ChainID] = ch
	return ch
}

// AddToLedger adds (synchronously confirms) transaction to the UTXODB ledger. Return error if it is
// invalid or double spend
func (env *Solo) AddToLedger(tx *iotago.Transaction) error {
	return env.utxoDB.AddToLedger(tx)
}

// RequestsForChain parses the transaction and returns all requests contained in it which have chainID as the target
func (env *Solo) RequestsForChain(tx *iotago.Transaction, chainID isc.ChainID) ([]isc.Request, error) {
	env.chainsMutex.RLock()
	defer env.chainsMutex.RUnlock()

	m := env.requestsByChain(tx)
	ret, ok := m[chainID]
	if !ok {
		return nil, fmt.Errorf("chain %s does not exist", chainID.String())
	}
	return ret, nil
}

// requestsByChain parses the transaction and extracts those outputs which are interpreted as a request to a chain
func (env *Solo) requestsByChain(tx *iotago.Transaction) map[isc.ChainID][]isc.Request {
	ret, err := isc.RequestsInTransaction(tx)
	require.NoError(env.T, err)
	return ret
}

// AddRequestsToMempool adds all the requests to the chain mempool,
func (env *Solo) AddRequestsToMempool(ch *Chain, reqs []isc.Request) {
	ch.mempool.ReceiveRequests(reqs...)
}

// EnqueueRequests adds requests contained in the transaction to mempools of respective target chains
func (env *Solo) EnqueueRequests(tx *iotago.Transaction) {
	env.chainsMutex.RLock()
	defer env.chainsMutex.RUnlock()

	requests := env.requestsByChain(tx)

	for chainID, reqs := range requests {
		ch, ok := env.chains[chainID]
		if !ok {
			env.logger.Infof("dispatching requests. Unknown chain: %s", chainID.String())
			continue
		}
		ch.mempool.ReceiveRequests(reqs...)
	}
}

func (ch *Chain) GetAnchorOutputFromL1() *isc.AliasOutputWithID {
	outputs := ch.Env.utxoDB.GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(ch.Env.T, 1, len(outputs))
	for outputID, aliasOutput := range outputs {
		return isc.NewAliasOutputWithID(aliasOutput, outputID)
	}
	panic("unreachable")
}

// collateBatch selects requests which are not time locked
// returns batch and and 'remains unprocessed' flag
func (ch *Chain) collateBatch() []isc.Request {
	// emulating variable sized blocks
	maxBatch := MaxRequestsInBlock - rand.Intn(MaxRequestsInBlock/3)
	requests := ch.mempool.RequestBatchProposal()
	batchSize := len(requests)

	if batchSize > maxBatch {
		batchSize = maxBatch
	}

	return requests[:batchSize]
}

func (ch *Chain) collateAndRunBatch() {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()
	if ch.Env.ctx.Err() != nil {
		return
	}
	batch := ch.collateBatch()
	if len(batch) > 0 {
		results := ch.runRequestsNolock(batch, "batchLoop")
		for _, res := range results {
			if res.Receipt.Error != nil {
				ch.log.Errorf("runRequestsSync: %v", res.Receipt.Error)
			}
		}
	}
}

func (ch *Chain) AddMigration(m migrations.Migration) {
	ch.migrationScheme.Migrations = append(ch.migrationScheme.Migrations, m)
}

func (ch *Chain) GetCandidateNodes() []*governance.AccessNodeInfo {
	panic("unimplemented")
}

func (ch *Chain) GetChainNodes() []peering.PeerStatusProvider {
	panic("unimplemented")
}

func (ch *Chain) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("unimplemented")
}

func (ch *Chain) ID() isc.ChainID {
	return ch.ChainID
}

func (ch *Chain) Log() *logger.Logger {
	return ch.log
}

func (ch *Chain) Processors() *processors.Cache {
	return ch.proc
}

func (ch *Chain) EnqueueDismissChain(_ string) {
	panic("unimplemented")
}

func (ch *Chain) EnqueueAliasOutput(_ *isc.AliasOutputWithID) {
	panic("unimplemented")
}

// ---------------------------------------------

func (env *Solo) UnspentOutputs(addr iotago.Address) (iotago.OutputSet, iotago.OutputIDs) {
	allOuts, _ := env.utxoDB.GetUnspentOutputs(addr)
	ids := make(iotago.OutputIDs, len(allOuts))
	i := 0
	for id := range allOuts {
		ids[i] = id
		i++
	}
	return allOuts, ids
}

func (env *Solo) L1NFTs(addr iotago.Address) map[iotago.OutputID]*iotago.NFTOutput {
	return env.utxoDB.GetAddressNFTs(addr)
}

// L1NativeTokens returns number of native tokens contained in the given address on the UTXODB ledger
func (env *Solo) L1NativeTokens(addr iotago.Address, nativeTokenID iotago.NativeTokenID) *big.Int {
	assets := env.L1Assets(addr)
	return assets.AmountNativeToken(nativeTokenID)
}

func (env *Solo) L1BaseTokens(addr iotago.Address) uint64 {
	return env.utxoDB.GetAddressBalances(addr).BaseTokens
}

// L1Assets returns all ftokens of the address contained in the UTXODB ledger
func (env *Solo) L1Assets(addr iotago.Address) *isc.Assets {
	return env.utxoDB.GetAddressBalances(addr)
}

func (env *Solo) L1Ledger() *utxodb.UtxoDB {
	return env.utxoDB
}

type NFTMintedInfo struct {
	Output   iotago.Output
	OutputID iotago.OutputID
	NFTID    iotago.NFTID
}

// MintNFTL1 mints a single NFT with the `issuer` account and sends it to a `target` account.
// Base tokens in the NFT output are sent to the minimum storage deposit and are taken from the issuer account.
func (env *Solo) MintNFTL1(issuer *cryptolib.KeyPair, target iotago.Address, immutableMetadata []byte) (*isc.NFT, *NFTMintedInfo, error) {
	nfts, infos, err := env.MintNFTsL1(issuer, target, nil, [][]byte{immutableMetadata})
	if err != nil {
		return nil, nil, err
	}
	return nfts[0], infos[0], nil
}

// MintNFTsL1 mints len(immutableMetadata) NFTs with the `issuer` account and sends them
// to a `target` account.
//
// If collectionOutputID is not nil, it must be an outputID of an NFTOutput owned by the issuer.
// All minted NFTs will belong to the given collection.
// See: https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
//
// Base tokens in the NFT outputs are sent to the minimum storage deposit and are taken from the issuer account.
func (env *Solo) MintNFTsL1(issuer *cryptolib.KeyPair, target iotago.Address, collectionOutputID *iotago.OutputID, immutableMetadata [][]byte) ([]*isc.NFT, []*NFTMintedInfo, error) {
	allOuts, allOutIDs := env.utxoDB.GetUnspentOutputs(issuer.Address())

	tx, err := transaction.NewMintNFTsTransaction(transaction.MintNFTsTransactionParams{
		IssuerKeyPair:      issuer,
		CollectionOutputID: collectionOutputID,
		Target:             target,
		ImmutableMetadata:  immutableMetadata,
		UnspentOutputs:     allOuts,
		UnspentOutputIDs:   allOutIDs,
	})
	if err != nil {
		return nil, nil, err
	}
	err = env.AddToLedger(tx)
	if err != nil {
		return nil, nil, err
	}

	outSet, err := tx.OutputsSet()
	if err != nil {
		return nil, nil, err
	}

	var nfts []*isc.NFT
	var infos []*NFTMintedInfo
	for id, out := range outSet {
		if out, ok := out.(*iotago.NFTOutput); ok { //nolint:gocritic // false positive
			nftID := util.NFTIDFromNFTOutput(out, id)
			info := &NFTMintedInfo{
				OutputID: id,
				Output:   out,
				NFTID:    nftID,
			}
			nft := &isc.NFT{
				ID:       info.NFTID,
				Issuer:   out.ImmutableFeatureSet().IssuerFeature().Address,
				Metadata: out.ImmutableFeatureSet().MetadataFeature().Data,
			}
			nfts = append(nfts, nft)
			infos = append(infos, info)
		}
	}
	return nfts, infos, nil
}

// SendL1 sends base or native tokens to another L1 address
func (env *Solo) SendL1(targetAddress iotago.Address, assets *isc.Assets, wallet *cryptolib.KeyPair) {
	allOuts, allOutIDs := env.utxoDB.GetUnspentOutputs(wallet.Address())
	tx, err := transaction.NewTransferTransaction(transaction.NewTransferTransactionParams{
		DisableAutoAdjustStorageDeposit: env.disableAutoAdjustStorageDeposit,
		FungibleTokens:                  assets,
		SendOptions:                     isc.SendOptions{},
		SenderAddress:                   wallet.Address(),
		SenderKeyPair:                   wallet,
		TargetAddress:                   targetAddress,
		UnspentOutputs:                  allOuts,
		UnspentOutputIDs:                allOutIDs,
	})
	require.NoError(env.T, err)
	err = env.AddToLedger(tx)
	require.NoError(env.T, err)
}
