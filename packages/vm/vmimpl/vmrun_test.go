package vmimpl

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/suitest"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// TODO
// func TestNFTDepositNoIssuer(t *testing.T) {
// 	metadata := isc.RequestMetadata{Message: accounts.FuncDeposit.Message()}
// 	o := &iotago.NFTOutput{
// 		Amount:       100 * isc.Million,
// 		NativeTokens: []*isc.NativeToken{},
// 		NFTID:        sui.ObjectID{0x1},
// 		Conditions:   []iotago.UnlockCondition{},
// 		Features: []iotago.Feature{
// 			&iotago.MetadataFeature{
// 				Data: metadata.Bytes(),
// 			},
// 			&iotago.SenderFeature{
// 				Address: tpkg.RandEd25519Address(),
// 			},
// 		},
// 		ImmutableFeatures: []iotago.Feature{
// 			&iotago.MetadataFeature{
// 				Data: []byte("foobar"),
// 			},
// 		},
// 	}
//
// 	res := simulateRunOutput(t, o)
// 	require.Len(t, res.RequestResults, 1)
// 	require.Nil(t, res.RequestResults[0].Receipt.Error)
// }

var schemaVersion = allmigrations.DefaultScheme.LatestSchemaVersion()

// initChain initializes a new chain state on the given empty store, and returns a fake L1
// anchor with a random ObjectID and the corresponding StateMetadata.
func initChain(chainCreator *cryptolib.KeyPair, store state.Store) *isc.StateAnchor {
	baseTokenCoinInfo := &isc.SuiCoinInfo{CoinType: coin.BaseTokenType}
	// create the anchor for a new chain
	initParams := origin.EncodeInitParams(
		isc.NewAddressAgentID(chainCreator.Address()),
		evm.DefaultChainID,
		governance.DefaultBlockKeepAmount,
	)
	const originDeposit = 1 * isc.Million
	_, stateMetadata := origin.InitChain(
		schemaVersion,
		store,
		initParams,
		originDeposit,
		baseTokenCoinInfo,
	)
	stateMetadataBytes := stateMetadata.Bytes()
	anchor := iscmove.Anchor{
		ID:            *suitest.RandomAddress(),
		StateMetadata: stateMetadataBytes,
		StateIndex:    0,
		Assets: iscmove.AssetsBag{
			ID:   *suitest.RandomAddress(),
			Size: 1,
		},
	}
	return &isc.StateAnchor{
		Anchor: &iscmove.AnchorWithRef{
			ObjectRef: sui.ObjectRef{
				ObjectID: &anchor.ID,
				Version:  0,
			},
			Object: &anchor,
		},
		Owner: chainCreator.Address(),
	}
}

// makeOnLedgerRequest creates a fake OnLedgerRequest
func makeOnLedgerRequest(
	t *testing.T,
	sender *cryptolib.KeyPair,
	chainID isc.ChainID,
	msg isc.Message,
	baseTokens uint64,
) isc.OnLedgerRequest {
	requestRef := suitest.RandomObjectRef()
	requestAssetsBagRef := suitest.RandomObjectRef()
	request := &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *requestRef,
		Object: &iscmove.Request{
			ID:     *requestRef.ObjectID,
			Sender: sender.Address(),
			AssetsBag: iscmove.AssetsBagWithBalances{
				AssetsBag: iscmove.AssetsBag{
					ID:   *requestAssetsBagRef.ObjectID,
					Size: 1,
				},
				Balances: map[string]*suijsonrpc.Balance{
					string(coin.BaseTokenType): {
						CoinType:        string(coin.BaseTokenType),
						CoinObjectCount: 1,
						TotalBalance:    baseTokens,
					},
				},
			},
			Message: iscmove.Message{
				Contract: uint32(msg.Target.Contract),
				Function: uint32(msg.Target.EntryPoint),
				Args:     msg.Params,
			},
			Allowance: []iscmove.CoinAllowance{},
			GasBudget: 1000,
		},
	}
	req, err := isc.OnLedgerFromRequest(request, chainID.AsAddress())
	require.NoError(t, err)
	return req
}

// transitionAnchor creates a new version of the Anchor given the result of the VM run
func transitionAnchor(
	t *testing.T,
	anchor *isc.StateAnchor,
	store indexedstore.IndexedStore,
	block state.Block,
) *isc.StateAnchor {
	require.EqualValues(t, anchor.Anchor.Version+1, block.StateIndex())

	stateMetadata := lo.Must(transaction.StateMetadataFromBytes(anchor.Anchor.Object.StateMetadata))

	state := lo.Must(store.StateByTrieRoot(block.TrieRoot()))
	chainInfo := governance.NewStateReaderFromChainState(state).
		GetChainInfo(isc.ChainIDFromObjectID(*anchor.Anchor.ObjectID))
	allCoinBalances := accounts.NewStateReaderFromChainState(stateMetadata.SchemaVersion, state).
		GetTotalL2FungibleTokens()

	newStateMetadata := transaction.NewStateMetadata(
		stateMetadata.SchemaVersion,
		block.L1Commitment(),
		chainInfo.GasFeePolicy,
		stateMetadata.InitParams,
		chainInfo.PublicURL,
	)
	return &isc.StateAnchor{
		Anchor: &iscmove.AnchorWithRef{
			ObjectRef: sui.ObjectRef{
				ObjectID: anchor.Anchor.ObjectID,
				Version:  anchor.Anchor.Version + 1,
			},
			Object: &iscmove.Anchor{
				ID: *anchor.Anchor.ObjectID,
				Assets: iscmove.AssetsBag{
					ID:   anchor.Anchor.Object.Assets.ID,
					Size: uint64(len(allCoinBalances)),
				},
				StateMetadata: newStateMetadata.Bytes(),
				StateIndex:    block.StateIndex(),
			},
		},
		Owner: anchor.Owner,
	}
}

func runRequestsAndTransitionAnchor(
	t *testing.T,
	anchor *isc.StateAnchor,
	store indexedstore.IndexedStore,
	reqs []isc.Request,
) (
	state.Block,
	*isc.StateAnchor,
) {
	task := &vm.VMTask{
		Processors:           processors.MustNew(coreprocessors.NewConfigWithCoreContracts()),
		Anchor:               anchor,
		Store:                store,
		Requests:             reqs,
		Timestamp:            time.Time{},
		Entropy:              [32]byte{},
		ValidatorFeeTarget:   nil,
		EstimateGasMode:      false,
		EVMTracer:            nil,
		EnableGasBurnLogging: false,
		Migrations:           allmigrations.DefaultScheme,
		Log:                  testlogger.NewLogger(t),
	}
	res, err := Run(task)
	require.NoError(t, err)
	require.Len(t, res.RequestResults, 1)
	require.Nil(t, res.RequestResults[0].Receipt.Error)
	require.NotNil(t, res.UnsignedTransaction)
	require.NotNil(t, res.StateDraft)

	block := store.Commit(res.StateDraft)
	store.SetLatest(block.TrieRoot())
	anchor = transitionAnchor(t, anchor, store, block)
	return block, anchor
}

func TestOnLedgerAccountsDeposit(t *testing.T) {
	chainCreator := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("chainCreator")))
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	anchor := initChain(chainCreator, store)
	chainID := isc.ChainIDFromObjectID(*anchor.Anchor.ObjectID)

	sender := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("sender")))
	{
		state := lo.Must(store.LatestState())
		senderL2Balance := accounts.NewStateReaderFromChainState(schemaVersion, state).
			GetAccountFungibleTokens(isc.NewAddressAgentID(sender.Address()), chainID)
		require.Zero(t, senderL2Balance.BaseTokens())
	}

	const baseTokens = 1 * isc.Million
	req := makeOnLedgerRequest(
		t,
		sender,
		chainID,
		accounts.FuncDeposit.Message(),
		baseTokens,
	)

	block, nextAnchor := runRequestsAndTransitionAnchor(
		t,
		anchor,
		store,
		[]isc.Request{req},
	)
	require.Equal(t, block.StateIndex(), nextAnchor.Anchor.Object.StateIndex)

	{
		state := lo.Must(store.LatestState())
		require.Equal(t, block.StateIndex(), state.BlockIndex())
		senderL2Balance := accounts.NewStateReaderFromChainState(schemaVersion, state).
			GetAccountFungibleTokens(isc.NewAddressAgentID(sender.Address()), chainID)
		receipt := lo.Must(blocklog.NewStateReaderFromChainState(state).
			GetRequestReceipt(req.ID()))
		require.EqualValues(t, baseTokens-receipt.GasFeeCharged, senderL2Balance.BaseTokens())
	}
}
