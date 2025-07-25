// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"context"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/evm/evmlogger"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/publisher"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
	_ "github.com/iotaledger/wasp/v2/packages/vm/sandbox"
)

const (
	timeLayout = "04:05.000000000"
)

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T                    Context
	logger               log.Logger
	chainsMutex          sync.RWMutex
	chains               map[isc.ChainID]*Chain
	processorConfig      *processors.Config
	enableGasBurnLogging bool
	seed                 cryptolib.Seed
	publisher            *publisher.Publisher
	ctx                  context.Context
	mockTime             time.Time
	l1ParamsFetcher      parameters.L1ParamsFetcher

	l1Config L1Config
}

// data to be persisted in the snapshot
type chainData struct {
	// Name is the name of the chain
	Name string

	// ChainID is the ID of the chain (in this version alias of the ChainAddress)
	ChainID isc.ChainID

	// AnchorOwner the key pair used to create and operate the chain.
	AnchorOwner *cryptolib.KeyPair

	// ChainAdmin the key pair designed as chain admin
	ChainAdmin *cryptolib.KeyPair

	db kvstore.KVStore

	migrationScheme *migrations.MigrationScheme
}

// Chain represents state of individual chain.
// There may be several parallel instances of the chain in the 'solo' test
type Chain struct {
	chainData

	// Env is a pointer to the global structure of the 'solo' test
	Env *Solo

	// Store is where the chain data (blocks, state) is stored
	store indexedstore.IndexedStore
	// Log is the named logger of the chain
	log log.Logger
	// global processor cache
	proc *processors.Config
	// related to asynchronous backlog processing
	runVMMutex sync.Mutex

	migrationScheme *migrations.MigrationScheme
}

type InitOptions struct {
	L1Config          *L1Config
	Debug             bool
	PrintStackTrace   bool
	GasBurnLogEnabled bool
	Log               log.Logger
}

type L1Config struct {
	IotaRPCURL    string
	IotaFaucetURL string
	ISCPackageID  iotago.PackageID
}

func DefaultInitOptions() *InitOptions {
	return &InitOptions{
		Debug:             false,
		PrintStackTrace:   false,
		GasBurnLogEnabled: true, // is ON by default
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
			opt.Log = testlogger.WithLevel(opt.Log, log.LevelInfo, opt.PrintStackTrace)
		}
	}
	evmlogger.Init(opt.Log)

	if opt.L1Config == nil {
		opt.L1Config = &L1Config{
			IotaRPCURL:    l1starter.Instance().APIURL(),
			IotaFaucetURL: l1starter.Instance().FaucetURL(),
			ISCPackageID:  l1starter.Instance().ISCPackageID(),
		}
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	t.Cleanup(cancelCtx)

	ret := &Solo{
		T:                    t,
		logger:               opt.Log,
		l1Config:             *opt.L1Config,
		chains:               make(map[isc.ChainID]*Chain),
		processorConfig:      coreprocessors.NewConfigWithTestContracts(),
		enableGasBurnLogging: opt.GasBurnLogEnabled,
		seed:                 cryptolib.NewSeed(),
		publisher:            publisher.New(opt.Log.NewChildLogger("publisher")),
		l1ParamsFetcher:      parameters.NewL1ParamsFetcher(l1starter.Instance().L1Client().IotaClient(), opt.Log),
		ctx:                  ctx,
	}
	_ = ret.publisher.Events.Published.Hook(func(ev *publisher.ISCEvent[any]) {
		ret.logger.LogInfof("solo publisher: %s %s %v", ev.Kind, ev.ChainID, ev.String())
	})

	go ret.publisher.Run(ctx)

	return ret
}

func (env *Solo) IterateChainTrieDBs(
	f func(chainID *isc.ChainID, k []byte, v []byte),
) {
	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()

	chainIDs := lo.Keys(env.chains)
	slices.SortFunc(chainIDs, func(a, b isc.ChainID) int { return bytes.Compare(a.Bytes(), b.Bytes()) })
	for _, chID := range chainIDs {
		ch := env.chains[chID]
		lo.Must0(ch.db.Iterate(nil, func(k []byte, v []byte) bool {
			f(&chID, k, v)
			return true
		}))
	}
}

func (env *Solo) IterateChainLatestStates(
	prefix kv.Key,
	f func(chainID *isc.ChainID, k []byte, v []byte),
) {
	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()

	chainIDs := lo.Keys(env.chains)
	slices.SortFunc(chainIDs, func(a, b isc.ChainID) int { return bytes.Compare(a.Bytes(), b.Bytes()) })
	for _, chID := range chainIDs {
		ch := env.chains[chID]
		store := indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(ch.db))
		state, err := store.LatestState()
		require.NoError(env.T, err)
		state.IterateSorted(prefix, func(k kv.Key, v []byte) bool {
			f(&chID, []byte(k), v)
			return true
		})
	}
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

const (
	DefaultChainAdminBaseTokens = 50 * isc.Million
)

// NewChain deploys a new default chain instance.
func (env *Solo) NewChain(depositFundsForAdmin ...bool) *Chain {
	ret, _ := env.NewChainExt(nil, 0, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	if len(depositFundsForAdmin) == 0 || depositFundsForAdmin[0] {
		// deposit some tokens for the chain originator
		err := ret.DepositBaseTokensToL2(DefaultChainAdminBaseTokens, ret.ChainAdmin)
		require.NoError(env.T, err)
	}
	return ret
}

func (env *Solo) ISCPackageID() iotago.PackageID {
	return env.l1Config.ISCPackageID
}

// MustWithWaitForNextVersion waits for an object to change its version and panics on timeouts
// This tries to make sure that an object meant to be used multiple times, does not get referenced twice with the same ref.
// Handle with care. Only use it on objects that are expected to be used again, like a GasCoin/Generic coin/Requests
func (env *Solo) MustWithWaitForNextVersion(currentRef *iotago.ObjectRef, cb func()) *iotago.ObjectRef {
	return lo.Must(env.WithWaitForNextVersion(currentRef, cb))
}

// WithWaitForNextVersion waits for an object to change its version.
// This tries to make sure that an object meant to be used multiple times, does not get referenced twice with the same ref.
// Handle with care. Only use it on objects that are expected to be used again, like a GasCoin/Generic coin/Requests
func (env *Solo) WithWaitForNextVersion(currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	return env.L1Client().WaitForNextVersionForTesting(context.Background(), 30*time.Second, env.logger, currentRef, cb)
}

func (env *Solo) deployChain(chainAdmin *cryptolib.KeyPair, initCommonAccountBaseTokens coin.Value, name string, evmChainID uint16, blockKeepAmount int32) chainData {
	env.logger.LogDebugf("deploying new chain '%s'", name)

	if chainAdmin == nil {
		chainAdmin = env.NewKeyPairFromIndex(-1000 + len(env.chains)) // making new originator for each new chain
		env.GetFundsFromFaucet(chainAdmin.Address())
	}

	anchorOwner := env.NewKeyPairFromIndex(-2000 + len(env.chains))
	env.GetFundsFromFaucet(anchorOwner.Address())

	initParams := origin.NewInitParams(
		isc.NewAddressAgentID(chainAdmin.Address()),
		evmChainID,
		blockKeepAmount,
		true,
	)

	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	db := mapdb.NewMapDB()
	store := indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(db))

	gasCoinRef := env.makeBaseTokenCoin(anchorOwner, isc.GasCoinTargetValue, nil)
	env.logger.LogInfof("Chain Originator address: %v\n", anchorOwner)
	env.logger.LogInfof("GAS COIN BEFORE PULL: %v\n", gasCoinRef)

	var block state.Block
	var stateMetadata *transaction.StateMetadata

	block, stateMetadata = origin.InitChain(
		schemaVersion,
		store,
		initParams.Encode(),
		*gasCoinRef.ObjectID,
		initCommonAccountBaseTokens,
		env.L1Params(),
	)

	var initCoin *iotago.ObjectRef

	if initCommonAccountBaseTokens > 0 {
		initCoin = env.makeBaseTokenCoin(
			anchorOwner,
			initCommonAccountBaseTokens,
			func(c *iotajsonrpc.Coin) bool {
				return !c.CoinObjectID.Equals(*gasCoinRef.ObjectID)
			},
		)
	}

	gasPayment, err := iotajsonrpc.PickupCoinsWithFilter(
		env.L1BaseTokenCoins(anchorOwner.Address()),
		uint64(iotaclient.DefaultGasBudget),
		func(c *iotajsonrpc.Coin) bool {
			return !c.CoinObjectID.Equals(*gasCoinRef.ObjectID) &&
				(initCoin == nil || !c.CoinObjectID.Equals(*initCoin.ObjectID))
		},
	)
	require.NoError(env.T, err)

	var anchorRef *iscmove.AnchorWithRef
	env.MustWithWaitForNextVersion(gasPayment.CoinRefs()[0], func() {
		env.MustWithWaitForNextVersion(initCoin, func() {
			anchorRef, err = env.ISCMoveClient().StartNewChain(
				env.ctx,
				&iscmoveclient.StartNewChainRequest{
					Signer:        anchorOwner,
					AnchorOwner:   anchorOwner.Address(),
					PackageID:     env.ISCPackageID(),
					StateMetadata: stateMetadata.Bytes(),
					InitCoinRef:   initCoin,
					GasPrice:      iotaclient.DefaultGasPrice,
					GasBudget:     iotaclient.DefaultGasBudget,
					GasPayments:   gasPayment.CoinRefs(),
				},
			)
		})
	})

	require.NoError(env.T, err)
	chainID := isc.ChainIDFromObjectID(anchorRef.Object.ID)

	env.logger.LogInfof(
		"deployed chain '%s' - ID: %s - anchor owner: %s - chain admin: %s - origin trie root: %s",
		name,
		chainID,
		anchorOwner.Address(),
		chainAdmin.Address(),
		block.TrieRoot(),
	)

	return chainData{
		Name:            name,
		ChainID:         chainID,
		AnchorOwner:     anchorOwner,
		ChainAdmin:      chainAdmin,
		db:              db,
		migrationScheme: allmigrations.DefaultScheme,
	}
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
	initCommonAccountBaseTokens coin.Value,
	name string,
	evmChainID uint16,
	blockKeepAmount int32,
) (*Chain, *isc.StateAnchor) {
	chData := env.deployChain(
		chainOriginator,
		initCommonAccountBaseTokens,
		name,
		evmChainID,
		blockKeepAmount,
	)

	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()
	ch := env.addChain(chData)

	ch.log.LogInfof("chain '%s' deployed. Chain ID: %s", ch.Name, ch.ChainID.String())
	return ch, nil
}

func (env *Solo) addChain(chData chainData) *Chain {
	ch := &Chain{
		chainData:       chData,
		Env:             env,
		store:           indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(chData.db)),
		proc:            env.processorConfig,
		log:             env.logger.NewChildLogger(chData.Name),
		migrationScheme: chData.migrationScheme,
	}
	env.chains[chData.ChainID] = ch
	return ch
}

func (env *Solo) Ctx() context.Context {
	return env.ctx
}

func (env *Solo) IotaFaucetURL() string {
	return env.l1Config.IotaFaucetURL
}

func (ch *Chain) GetAnchor(stateIndex uint32) (*isc.StateAnchor, error) {
	anchor, err := ch.Env.ISCMoveClient().GetPastAnchorFromObjectID(
		ch.Env.ctx,
		ch.ChainID.AsAddress().AsIotaAddress(),
		uint64(stateIndex),
	)
	if err != nil {
		return nil, err
	}

	stateAnchor := isc.NewStateAnchor(anchor, ch.Env.ISCPackageID())
	return &stateAnchor, nil
}

func (ch *Chain) GetLatestAnchor() *isc.StateAnchor {
	anchor, err := ch.Env.ISCMoveClient().GetAnchorFromObjectID(
		ch.Env.ctx,
		ch.ChainID.AsAddress().AsIotaAddress(),
	)
	require.NoError(ch.Env.T, err)

	stateAnchor := isc.NewStateAnchor(anchor, ch.Env.ISCPackageID())
	return &stateAnchor
}

func (env *Solo) GetCoin(id *iotago.ObjectID) *coin.CoinWithRef {
	getObjRes, err := env.ISCMoveClient().GetObject(
		env.ctx,
		iotaclient.GetObjectRequest{
			ObjectID: id,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
		},
	)
	require.NoError(env.T, err)
	require.Nil(env.T, getObjRes.Error)
	var moveGasCoin iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &moveGasCoin)
	require.NoError(env.T, err)
	gasCoinRef := getObjRes.Data.Ref()
	return &coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(moveGasCoin.Balance),
		Ref:   &gasCoinRef,
	}
}

func (ch *Chain) GetLatestGasCoin() *coin.CoinWithRef {
	anchor, err := ch.Env.ISCMoveClient().GetAnchorFromObjectID(
		ch.Env.ctx,
		ch.ChainID.AsAddress().AsIotaAddress(),
	)
	require.NoError(ch.Env.T, err)

	metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
	require.NoError(ch.Env.T, err)
	return ch.Env.GetCoin(metadata.GasCoinObjectID)
}

func (ch *Chain) GetLatestAnchorWithBalances() (*isc.StateAnchor, *isc.Assets) {
	anchor := ch.GetLatestAnchor()
	bals, err := ch.Env.ISCMoveClient().GetAssetsBagWithBalances(ch.Env.ctx, &anchor.GetAssetsBag().ID)
	require.NoError(ch.Env.T, err)
	return anchor, lo.Must(isc.AssetsFromAssetsBagWithBalances(bals))
}

// collateBatch selects requests to be processed in a batch
func (ch *Chain) collateBatch(maxRequestsInBlock int) []isc.Request {
	reqs := make([]*iscmove.RefWithObject[iscmove.Request], 0)
	err := ch.Env.ISCMoveClient().GetRequestsSorted(ch.Env.ctx, ch.Env.ISCPackageID(), ch.ChainID.AsAddress().AsIotaAddress(), maxRequestsInBlock, func(err error, i *iscmove.RefWithObject[iscmove.Request]) {
		require.NoError(ch.Env.T, err)
		reqs = append(reqs, i)
	})
	require.NoError(ch.Env.T, err)
	return lo.Map(reqs, func(req *iscmove.RefWithObject[iscmove.Request], _ int) isc.Request {
		r, err := isc.OnLedgerFromMoveRequest(req, ch.ChainID.AsAddress())
		require.NoError(ch.Env.T, err)
		return r
	})
}

// RunRequestBatch runs a batch of requests pending to be processed
func (ch *Chain) RunRequestBatch(maxRequestsInBlock int) (
	*iotajsonrpc.IotaTransactionBlockResponse,
	[]*vm.RequestResult,
) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	batch := ch.collateBatch(maxRequestsInBlock)
	if len(batch) == 0 {
		return nil, nil // no requests to process
	}
	ptbRes, results := ch.runRequestsNolock(batch)
	for _, res := range results {
		if res.Receipt.Error != nil {
			ch.log.LogErrorf("runRequestsSync: %v", res.Receipt.Error)
		}
	}
	return ptbRes, results
}

func (ch *Chain) RunAllReceivedRequests(maxRequestsInBlock int) int {
	runs := 0
	for {
		_, res := ch.RunRequestBatch(maxRequestsInBlock)
		if res == nil {
			break
		}
		runs++
	}
	return runs
}

func (ch *Chain) AddMigration(m migrations.Migration) {
	ch.migrationScheme.Migrations = append(ch.migrationScheme.Migrations, m)
}

func (ch *Chain) ID() isc.ChainID {
	return ch.ChainID
}

func (ch *Chain) Log() log.Logger {
	return ch.log
}

func (ch *Chain) Processors() *processors.Config {
	return ch.proc
}

// ---------------------------------------------

func (env *Solo) L1CoinInfo(coinType coin.Type) *parameters.IotaCoinInfo {
	md, err := env.L1Client().GetCoinMetadata(env.ctx, coinType.String())
	require.NoError(env.T, err)
	ts, err := env.L1Client().GetTotalSupply(env.ctx, coinType.String())
	require.NoError(env.T, err)
	return parameters.IotaCoinInfoFromL1Metadata(coinType, md, coin.Value(ts.Value.Uint64()))
}

func (env *Solo) L1BaseTokenCoins(addr *cryptolib.Address) []*iotajsonrpc.Coin {
	return env.L1Coins(addr, coin.BaseTokenType)
}

func (env *Solo) L1AllCoins(addr *cryptolib.Address) iotajsonrpc.Coins {
	r, err := env.L1Client().GetCoins(env.ctx, iotaclient.GetCoinsRequest{
		Owner: addr.AsIotaAddress(),
		Limit: math.MaxUint,
	})
	require.NoError(env.T, err)
	return r.Data
}

func (env *Solo) L1Coins(addr *cryptolib.Address, coinType coin.Type) []*iotajsonrpc.Coin {
	coinTypeStr := coinType.String()
	r, err := env.L1Client().GetCoins(env.ctx, iotaclient.GetCoinsRequest{
		Owner:    addr.AsIotaAddress(),
		CoinType: &coinTypeStr,
		Limit:    math.MaxUint,
	})
	require.NoError(env.T, err)
	return r.Data
}

func (env *Solo) L1BaseTokens(addr *cryptolib.Address) coin.Value {
	return env.L1CoinBalance(addr, coin.BaseTokenType)
}

func (env *Solo) L1CoinBalance(addr *cryptolib.Address, coinType coin.Type) coin.Value {
	r, err := env.L1Client().GetBalance(env.ctx, iotaclient.GetBalanceRequest{
		Owner:    addr.AsIotaAddress(),
		CoinType: coinType.String(),
	})
	require.NoError(env.T, err)
	return coin.Value(r.TotalBalance.Uint64())
}

// L1CoinBalances returns all ftokens of the address contained in the UTXODB ledger
func (env *Solo) L1CoinBalances(addr *cryptolib.Address) isc.CoinBalances {
	r, err := env.L1Client().GetAllBalances(env.ctx, addr.AsIotaAddress())
	require.NoError(env.T, err)
	cb := isc.NewCoinBalances()
	for _, b := range r {
		cb.Add(lo.Must(coin.TypeFromString(b.CoinType.String())), coin.Value(b.TotalBalance.Uint64()))
	}
	return cb
}

func (env *Solo) executePTB(
	ptb iotago.ProgrammableTransaction,
	wallet *cryptolib.KeyPair,
	gasPaymentCoins []*iotago.ObjectRef,
	gasBudget, gasPrice uint64,
) *iotajsonrpc.IotaTransactionBlockResponse {
	tx := iotago.NewProgrammable(
		wallet.Address().AsIotaAddress(),
		ptb,
		gasPaymentCoins,
		gasBudget,
		gasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(env.T, err)

	execRes, err := env.L1Client().SignAndExecuteTransaction(
		env.ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      cryptolib.SignerToIotaSigner(wallet),
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:        true,
				ShowObjectChanges:  true,
				ShowEvents:         true,
				ShowInput:          true,
				ShowBalanceChanges: true,
				ShowRawEffects:     true,
				ShowRawInput:       true,
			},
		},
	)
	require.NoError(env.T, err)
	if !execRes.Effects.Data.IsSuccess() {
		env.T.Fatalf("PTB failed: %s", execRes.Effects.Data.V1.Status.Error)
	}
	return execRes
}

func (env *Solo) L1DeployCoinPackage(keyPair *cryptolib.KeyPair) (
	packageID *iotago.PackageID,
	treasuryCap *iotago.ObjectRef,
) {
	return iotaclienttest.DeployCoinPackage(
		env.T,
		env.L1Client().IotaClient(),
		cryptolib.SignerToIotaSigner(keyPair),
		contracts.Testcoin(),
	)
}

func (env *Solo) L1MintCoin(
	keyPair *cryptolib.KeyPair,
	packageID *iotago.PackageID,
	moduleName iotago.Identifier,
	typeTag iotago.Identifier,
	treasuryCapObject *iotago.ObjectRef,
	mintAmount uint64,
) (coinRef *iotago.ObjectRef) {
	return iotaclienttest.MintCoins(
		env.T,
		env.L1Client().IotaClient(),
		cryptolib.SignerToIotaSigner(keyPair),
		packageID,
		moduleName,
		typeTag,
		treasuryCapObject,
		mintAmount,
	)
}

func (env *Solo) L1MintObject(owner *cryptolib.KeyPair) isc.IotaObject {
	// Create a 2nd chain just to have a L1 object that we can deposit (the anchor)
	testAnchor, err := env.ISCMoveClient().StartNewChain(env.Ctx(), &iscmoveclient.StartNewChainRequest{
		GasBudget:     iotaclient.DefaultGasBudget,
		Signer:        owner,
		PackageID:     env.ISCPackageID(),
		StateMetadata: []byte{},
		AnchorOwner:   owner.Address(),
		InitCoinRef:   nil,
		GasPrice:      iotaclient.DefaultGasPrice,
	})
	require.NoError(env.T, err)

	o, err := env.ISCMoveClient().GetObject(env.Ctx(), iotaclient.GetObjectRequest{
		ObjectID: testAnchor.ObjectID,
		Options: &iotajsonrpc.IotaObjectDataOptions{
			ShowType: true,
		},
	})
	require.NoError(env.T, err)
	typ, err := iotago.ObjectTypeFromString(*o.Data.Type)
	require.NoError(env.T, err)
	return isc.NewIotaObject(*testAnchor.ObjectID, typ)
}

func (env *Solo) L1Params() *parameters.L1Params {
	l1Params, err := env.l1ParamsFetcher.GetOrFetchLatest(env.ctx)
	require.NoError(env.T, err)
	return l1Params
}
