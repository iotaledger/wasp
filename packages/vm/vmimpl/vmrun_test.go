package vmimpl

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"
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
)

// TODO
// func TestNFTDepositNoIssuer(t *testing.T) {
// 	metadata := isc.RequestMetadata{Message: accounts.FuncDeposit.Message()}
// 	o := &iotago.NFTOutput{
// 		Amount:       100 * isc.Million,
// 		NativeTokens: []*isc.NativeToken{},
// 		NFTID:        iotago.ObjectID{0x1},
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
	// create the anchor for a new chain
	initParams := origin.NewInitParams(
		isc.NewAddressAgentID(chainCreator.Address()),
		evm.DefaultChainID,
		governance.DefaultBlockKeepAmount,
		false,
	).Encode()
	const originDeposit = 1 * isc.Million
	_, stateMetadata := origin.InitChain(
		schemaVersion,
		store,
		initParams,
		iotago.ObjectID{},
		originDeposit,
		parameterstest.L1Mock,
	)
	stateMetadataBytes := stateMetadata.Bytes()
	anchor := iscmove.Anchor{
		ID:            *iotatest.RandomAddress(),
		StateMetadata: stateMetadataBytes,
		StateIndex:    0,
		Assets: iscmove.Referent[iscmove.AssetsBag]{
			ID: *iotatest.RandomAddress(),
			Value: &iscmove.AssetsBag{
				ID:   *iotatest.RandomAddress(),
				Size: 1,
			},
		},
	}

	stateAnchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{
		ObjectRef: iotago.ObjectRef{
			ObjectID: &anchor.ID,
			Digest:   lo.Must(iotago.NewDigest("foo")),
			Version:  0,
		},
		Object: &anchor,
		Owner:  chainCreator.Address().AsIotaAddress(),
	}, iotago.ObjectID{})

	return &stateAnchor
}

// makeOnLedgerRequest creates a fake OnLedgerRequest
func makeOnLedgerRequest(
	t *testing.T,
	sender *cryptolib.KeyPair,
	chainID isc.ChainID,
	msg isc.Message,
	baseTokens uint64,
) isc.OnLedgerRequest {
	requestRef := iotatest.RandomObjectRef()
	requestAssetsBagRef := iotatest.RandomObjectRef()
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
				Assets: iscmove.Assets{
					Coins: iscmove.CoinBalances{
						iotajsonrpc.IotaCoinType: iotajsonrpc.CoinValue(baseTokens),
					},
					Objects: iscmove.ObjectCollection{},
				},
			},
			Message: iscmove.Message{
				Contract: uint32(msg.Target.Contract),
				Function: uint32(msg.Target.EntryPoint),
				Args:     msg.Params,
			},
			Allowance: iscmove.Assets{},
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
	require.EqualValues(t, anchor.GetStateIndex()+1, block.StateIndex())

	stateMetadata := lo.Must(transaction.StateMetadataFromBytes(anchor.GetStateMetadata()))

	state := lo.Must(store.StateByTrieRoot(block.TrieRoot()))
	chainInfo := governance.NewStateReaderFromChainState(state).
		GetChainInfo(anchor.ChainID())

	newStateMetadata := transaction.NewStateMetadata(
		stateMetadata.SchemaVersion,
		block.L1Commitment(),
		stateMetadata.GasCoinObjectID,
		chainInfo.GasFeePolicy,
		stateMetadata.InitParams,
		stateMetadata.InitDeposit,
		chainInfo.PublicURL,
	)

	return isctest.UpdateStateAnchor(anchor, newStateMetadata.Bytes())
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
		Processors: coreprocessors.NewConfigWithTestContracts(),
		Anchor:     anchor,
		GasCoin: &coin.CoinWithRef{
			Value: isc.GasCoinTargetValue,
			Type:  coin.BaseTokenType,
			Ref:   iotatest.RandomObjectRef(),
		},
		L1Params:             parameterstest.L1Mock,
		Store:                store,
		Requests:             reqs,
		Timestamp:            time.Time{},
		Entropy:              [32]byte{},
		ValidatorFeeTarget:   nil,
		EstimateGasMode:      false,
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
	chainID := anchor.ChainID()

	sender := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("sender")))
	{
		state := lo.Must(store.LatestState())
		senderL2Balance := accounts.NewStateReaderFromChainState(schemaVersion, state).
			GetAccountFungibleTokens(isc.NewAddressAgentID(sender.Address()))
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
	require.Equal(t, block.StateIndex(), nextAnchor.GetStateIndex())

	{
		state := lo.Must(store.LatestState())
		require.Equal(t, block.StateIndex(), state.BlockIndex())
		senderL2Balance := accounts.NewStateReaderFromChainState(schemaVersion, state).
			GetAccountFungibleTokens(isc.NewAddressAgentID(sender.Address()))
		receipt := lo.Must(blocklog.NewStateReaderFromChainState(state).
			GetRequestReceipt(req.ID()))
		require.EqualValues(t, baseTokens-receipt.GasFeeCharged, senderL2Balance.BaseTokens())
	}
}
