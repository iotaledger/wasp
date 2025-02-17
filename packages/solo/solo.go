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
	"go.uber.org/zap/zapcore"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmlogger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
)

const (
	timeLayout = "04:05.000000000"
)

// Solo is a structure which contains global parameters of the test: one per test instance
type Solo struct {
	// instance of the test
	T                    Context
	logger               *logger.Logger
	chainsMutex          sync.RWMutex
	chains               map[isc.ChainID]*Chain
	processorConfig      *processors.Config
	enableGasBurnLogging bool
	seed                 cryptolib.Seed
	publisher            *publisher.Publisher
	ctx                  context.Context
	mockTime             time.Time

	l1Config L1Config
}

// data to be persisted in the snapshot
type chainData struct {
	// Name is the name of the chain
	Name string

	// ChainID is the ID of the chain (in this version alias of the ChainAddress)
	ChainID isc.ChainID

	// OriginatorPrivateKey the key pair used to create the chain (origin transaction).
	// It is a default key pair in many of Solo calls which require private key.
	OriginatorPrivateKey *cryptolib.KeyPair

	db kvstore.KVStore

	migrationScheme *migrations.MigrationScheme
}

// Chain represents state of individual chain.
// There may be several parallel instances of the chain in the 'solo' test
type Chain struct {
	chainData

	OriginatorAddress *cryptolib.Address
	OriginatorAgentID isc.AgentID

	// Env is a pointer to the global structure of the 'solo' test
	Env *Solo

	// Store is where the chain data (blocks, state) is stored
	store indexedstore.IndexedStore
	// Log is the named logger of the chain
	log *logger.Logger
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
	Log               *logger.Logger
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
			opt.Log = testlogger.WithLevel(opt.Log, zapcore.InfoLevel, opt.PrintStackTrace)
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

	parameters.InitL1(*l1starter.Instance().L1Client().IotaClient(), opt.Log)

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
		publisher:            publisher.New(opt.Log.Named("publisher")),
		ctx:                  ctx,
	}
	_ = ret.publisher.Events.Published.Hook(func(ev *publisher.ISCEvent[any]) {
		ret.logger.Infof("solo publisher: %s %s %v", ev.Kind, ev.ChainID, ev.String())
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
		chID := chID // prevent loop variable aliasing
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
		chID := chID // prevent loop variable aliasing
		ch := env.chains[chID]
		store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(ch.db))
		state, err := store.LatestState()
		require.NoError(env.T, err)
		state.IterateSorted(prefix, func(k kv.Key, v []byte) bool {
			f(&chID, []byte(k), v)
			return true
		})
	}
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

const (
	DefaultChainOriginatorBaseTokens = 50 * isc.Million
)

// NewChain deploys a new default chain instance.
func (env *Solo) NewChain(depositFundsForOriginator ...bool) *Chain {
	ret, _ := env.NewChainExt(nil, 0, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	if len(depositFundsForOriginator) == 0 || depositFundsForOriginator[0] {
		// deposit some tokens for the chain originator
		err := ret.DepositBaseTokensToL2(DefaultChainOriginatorBaseTokens, ret.OriginatorPrivateKey)
		require.NoError(env.T, err)
	}
	return ret
}

func (env *Solo) ISCPackageID() iotago.PackageID {
	return env.l1Config.ISCPackageID
}

func (env *Solo) deployChain(
	chainOriginator *cryptolib.KeyPair,
	initCommonAccountBaseTokens coin.Value,
	name string,
	evmChainID uint16,
	blockKeepAmount int32,
) (chainData, *isc.StateAnchor) {
	env.logger.Debugf("deploying new chain '%s'", name)

	if chainOriginator == nil {
		chainOriginator = env.NewKeyPairFromIndex(-1000 + len(env.chains)) // making new originator for each new chain
		originatorAddr := chainOriginator.GetPublicKey().AsAddress()
		env.GetFundsFromFaucet(originatorAddr)
	}

	initParams := origin.NewInitParams(
		isc.NewAddressAgentID(chainOriginator.Address()),
		evmChainID,
		blockKeepAmount,
		true,
	)

	originatorAddr := chainOriginator.GetPublicKey().AsAddress()

	baseTokenCoinInfo := env.L1CoinInfo(coin.BaseTokenType)

	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	db := mapdb.NewMapDB()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(db))

	gasCoinRef := env.makeBaseTokenCoin(chainOriginator, isc.GasCoinTargetValue)

	block, stateMetadata := origin.InitChain(
		schemaVersion,
		store,
		initParams.Encode(),
		*gasCoinRef.ObjectID,
		initCommonAccountBaseTokens,
		baseTokenCoinInfo,
	)

	var initCoin *iotago.ObjectRef
	if initCommonAccountBaseTokens > 0 {
		initCoin = env.makeBaseTokenCoin(
			chainOriginator,
			initCommonAccountBaseTokens,
		)
	}

	anchorRef, err := env.ISCMoveClient().StartNewChain(
		env.ctx,
		&iscmoveclient.StartNewChainRequest{
			Signer:            chainOriginator,
			ChainOwnerAddress: chainOriginator.Address(),
			PackageID:         env.ISCPackageID(),
			StateMetadata:     stateMetadata.Bytes(),
			InitCoinRef:       initCoin,
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(env.T, err)
	chainID := isc.ChainIDFromObjectID(anchorRef.Object.ID)

	env.logger.Infof(
		"deployed chain '%s' - ID: %s - state controller address: %s - origin trie root: %s",
		name,
		chainID,
		originatorAddr,
		block.TrieRoot(),
	)

	return chainData{
		Name:                 name,
		ChainID:              chainID,
		OriginatorPrivateKey: chainOriginator,
		db:                   db,
		migrationScheme:      allmigrations.DefaultScheme,
	}, nil
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
	chData, anchorRef := env.deployChain(
		chainOriginator,
		initCommonAccountBaseTokens,
		name,
		evmChainID,
		blockKeepAmount,
	)

	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()
	ch := env.addChain(chData)

	ch.log.Infof("chain '%s' deployed. Chain ID: %s", ch.Name, ch.ChainID.String())
	return ch, anchorRef
}

func (env *Solo) addChain(chData chainData) *Chain {
	ch := &Chain{
		chainData:         chData,
		OriginatorAddress: chData.OriginatorPrivateKey.GetPublicKey().AsAddress(),
		OriginatorAgentID: isc.NewAddressAgentID(chData.OriginatorPrivateKey.GetPublicKey().AsAddress()),
		Env:               env,
		store:             indexedstore.New(state.NewStoreWithUniqueWriteMutex(chData.db)),
		proc:              env.processorConfig,
		log:               env.logger.Named(chData.Name),
		migrationScheme:   chData.migrationScheme,
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

func (ch *Chain) GetLatestGasCoin() *coin.CoinWithRef {
	anchor, err := ch.Env.ISCMoveClient().GetAnchorFromObjectID(
		ch.Env.ctx,
		ch.ChainID.AsAddress().AsIotaAddress(),
	)
	require.NoError(ch.Env.T, err)

	metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
	require.NoError(ch.Env.T, err)
	getObjRes, err := ch.Env.ISCMoveClient().GetObject(
		ch.Env.ctx,
		iotaclient.GetObjectRequest{
			ObjectID: metadata.GasCoinObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
		},
	)
	require.NoError(ch.Env.T, err)
	var moveGasCoin iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &moveGasCoin)
	require.NoError(ch.Env.T, err)
	gasCoinRef := getObjRes.Data.Ref()
	return &coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(moveGasCoin.Balance),
		Ref:   &gasCoinRef,
	}
}

func (ch *Chain) LatestL1Parameters() *parameters.L1Params {
	return parameters.L1()
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
		r, err := isc.OnLedgerFromRequest(req, ch.ChainID.AsAddress())
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
			ch.log.Errorf("runRequestsSync: %v", res.Receipt.Error)
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

func (ch *Chain) Log() *logger.Logger {
	return ch.log
}

func (ch *Chain) Processors() *processors.Config {
	return ch.proc
}

// ---------------------------------------------

func (env *Solo) L1CoinInfo(coinType coin.Type) *isc.IotaCoinInfo {
	md, err := env.IotaClient().GetCoinMetadata(env.ctx, coinType.String())
	require.NoError(env.T, err)
	ts, err := env.IotaClient().GetTotalSupply(env.ctx, coinType.String())
	require.NoError(env.T, err)
	return isc.IotaCoinInfoFromL1Metadata(coinType, md, coin.Value(ts.Value.Uint64()))
}

func (env *Solo) L1BaseTokenCoins(addr *cryptolib.Address) []*iotajsonrpc.Coin {
	return env.L1Coins(addr, coin.BaseTokenType)
}

func (env *Solo) L1AllCoins(addr *cryptolib.Address) iotajsonrpc.Coins {
	r, err := env.IotaClient().GetCoins(env.ctx, iotaclient.GetCoinsRequest{
		Owner: addr.AsIotaAddress(),
		Limit: math.MaxUint,
	})
	require.NoError(env.T, err)
	return r.Data
}

func (env *Solo) L1Coins(addr *cryptolib.Address, coinType coin.Type) []*iotajsonrpc.Coin {
	coinTypeStr := coinType.String()
	r, err := env.IotaClient().GetCoins(env.ctx, iotaclient.GetCoinsRequest{
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
	r, err := env.IotaClient().GetBalance(env.ctx, iotaclient.GetBalanceRequest{
		Owner:    addr.AsIotaAddress(),
		CoinType: coinType.String(),
	})
	require.NoError(env.T, err)
	return coin.Value(r.TotalBalance.Uint64())
}

func (env *Solo) L1NFTs(addr *cryptolib.Address) []iotago.ObjectID {
	panic("TODO")
}

// L1Assets returns all ftokens of the address contained in the UTXODB ledger
func (env *Solo) L1CoinBalances(addr *cryptolib.Address) isc.CoinBalances {
	r, err := env.IotaClient().GetAllBalances(env.ctx, addr.AsIotaAddress())
	require.NoError(env.T, err)
	cb := isc.NewCoinBalances()
	for _, b := range r {
		cb.Add(lo.Must(coin.TypeFromString(b.CoinType.String())), coin.Value(b.TotalBalance.Uint64()))
	}
	return cb
}

// MintNFTL1 mints a single NFT with the `issuer` account and sends it to a `target` account.
// Base tokens in the NFT output are sent to the minimum storage deposit and are taken from the issuer account.
func (env *Solo) MintNFTL1(issuer *cryptolib.KeyPair, target *cryptolib.Address, immutableMetadata []byte) (*isc.NFT, error) {
	nfts, err := env.MintNFTsL1(issuer, target, nil, [][]byte{immutableMetadata})
	if err != nil {
		return nil, err
	}
	return nfts[0], nil
}

// MintNFTsL1 mints len(metadata) NFTs with the `issuer` account and sends them
// to a `target` account.
//
// If collectionID is not nil, it must be the ID of an NFT owned by the issuer.
// All minted NFTs will belong to the given collection.
// See: https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
//
// Base tokens in the NFT outputs are sent to the minimum storage deposit and are taken from the issuer account.
func (env *Solo) MintNFTsL1(
	issuer *cryptolib.KeyPair,
	target *cryptolib.Address,
	collectionID *iotago.ObjectID,
	metadata [][]byte,
) ([]*isc.NFT, error) {
	panic("TODO")
	/*
		err := errors.New("refactor me: MintNFTsL1")
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
					Issuer:   cryptolib.NewAddressFromIotago(out.ImmutableFeatureSet().IssuerFeature().Address),
					Metadata: out.ImmutableFeatureSet().MetadataFeature().Data,
				}
				nfts = append(nfts, nft)
				infos = append(infos, info)
			}
		}
		return nfts, infos, nil
	*/
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

	execRes, err := env.IotaClient().SignAndExecuteTransaction(
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
	require.True(env.T, execRes.Effects.Data.IsSuccess())
	return execRes
}

func (env *Solo) L1DeployCoinPackage(keyPair *cryptolib.KeyPair) (
	packageID *iotago.PackageID,
	treasuryCap *iotago.ObjectRef,
) {
	return iotaclienttest.DeployCoinPackage(
		env.T,
		env.IotaClient(),
		cryptolib.SignerToIotaSigner(keyPair),
		contracts.Testcoin(),
	)
}

func (env *Solo) L1MintCoin(
	keyPair *cryptolib.KeyPair,
	packageID *iotago.PackageID,
	moduleName iotago.Identifier,
	typeTag iotago.Identifier,
	treasuryCapObjectID *iotago.ObjectID,
	mintAmount uint64,
) (coinRef *iotago.ObjectRef) {
	return iotaclienttest.MintCoins(
		env.T,
		env.IotaClient(),
		cryptolib.SignerToIotaSigner(keyPair),
		packageID,
		moduleName,
		typeTag,
		treasuryCapObjectID,
		mintAmount,
	)
}
