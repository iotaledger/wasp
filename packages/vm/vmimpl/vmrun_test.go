package vmimpl

import (
	"testing"
	"time"

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
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/sui-go/sui"
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

func initChain(chainCreator *cryptolib.KeyPair, store state.Store) *isc.StateAnchor {
	// create the anchor for a new chain
	initParams := origin.EncodeInitParams(
		isc.NewAddressAgentID(chainCreator.Address()),
		evm.DefaultChainID,
		governance.DefaultBlockKeepAmount,
	)
	const originDeposit = 1 * isc.Million
	_, stateMetadata := origin.InitChain(
		allmigrations.SchemaVersionIotaRebased,
		store,
		initParams,
		originDeposit,
		baseTokenCoinInfo,
	)
	anchorObjectRef := sui.RandomObjectRef()
	anchorAssetsBagRef := sui.RandomObjectRef()
	anchorAssetsReferentRef := sui.RandomObjectRef()
	anchor := &isc.StateAnchor{
		Ref: &iscmove.AnchorWithRef{
			ObjectRef: *anchorObjectRef,
			Object: &iscmove.Anchor{
				ID: *anchorObjectRef.ObjectID,
				Assets: iscmove.Referent[iscmove.AssetsBag]{
					ID: *anchorAssetsReferentRef.ObjectID,
					Value: &iscmove.AssetsBag{
						ID:   *anchorAssetsBagRef.ObjectID,
						Size: 0,
					},
				},
				StateMetadata: stateMetadata.Bytes(),
				StateIndex:    0,
			},
		},
		Owner: chainCreator.Address(),
	}
	return anchor
}

func TestRunVM(t *testing.T) {
	chainCreator := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("chainCreator")))
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	anchor := initChain(chainCreator, store)

	chainID := isc.ChainIDFromObjectID(*anchor.Ref.ObjectID)

	// create a request
	sender := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("sender")))
	requestRef := sui.RandomObjectRef()
	requestAssetsBagRef := sui.RandomObjectRef()
	requestAssetsReferentRef := sui.RandomObjectRef()
	msg := accounts.FuncDeposit.Message()
	const tokensForGas = 1 * isc.Million
	request := &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *requestRef,
		Object: &iscmove.Request{
			ID:     *requestRef.ObjectID,
			Sender: sender.Address(),
			AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
				ID: *requestAssetsReferentRef.ObjectID,
				Value: &iscmove.AssetsBagWithBalances{
					AssetsBag: iscmove.AssetsBag{
						ID:   *requestAssetsBagRef.ObjectID,
						Size: 1,
					},
					Balances: map[string]*suijsonrpc.Balance{
						string(coin.BaseTokenType): {
							CoinType:        string(coin.BaseTokenType),
							CoinObjectCount: 1,
							TotalBalance:    tokensForGas,
						},
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

	// create task and run it
	task := &vm.VMTask{
		Processors:           processors.MustNew(coreprocessors.NewConfigWithCoreContracts()),
		Anchor:               anchor,
		Store:                store,
		Requests:             []isc.Request{req},
		Timestamp:            time.Time{},
		Entropy:              [32]byte{},
		ValidatorFeeTarget:   nil,
		EstimateGasMode:      false,
		EVMTracer:            nil,
		EnableGasBurnLogging: false,
		Migrations:           allmigrations.DefaultScheme,
		Log:                  testlogger.NewLogger(t),
	}
	res := runTask(task)
	require.Len(t, res.RequestResults, 1)
	require.Nil(t, res.RequestResults[0].Receipt.Error)
	require.NotNil(t, res.UnsignedTransaction)
}
