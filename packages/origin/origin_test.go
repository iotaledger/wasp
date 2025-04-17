package origin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestOrigin(t *testing.T) {
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := origin.DefaultInitParams(isctest.NewRandomAgentID()).Encode()
	originDepositVal := coin.Value(100)
	l1commitment := origin.L1Commitment(schemaVersion, initParams, iotago.ObjectID{}, originDepositVal, parameterstest.L1Mock)
	store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	initBlock, _ := origin.InitChain(schemaVersion, store, initParams, iotago.ObjectID{}, originDepositVal, parameterstest.L1Mock)
	latestBlock, err := store.LatestBlock()
	require.NoError(t, err)
	require.True(t, l1commitment.Equals(initBlock.L1Commitment()))
	require.True(t, l1commitment.Equals(latestBlock.L1Commitment()))
}

func TestCreateOrigin(t *testing.T) {
	client := iscmoveclienttest.NewHTTPClient()
	sentSigner := iscmoveclienttest.NewRandomSignerWithFunds(t, 0)
	stateSigner := iscmoveclienttest.NewRandomSignerWithFunds(t, 1)
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(sentSigner.Address())).Encode()

	coinType := iotajsonrpc.IotaCoinType.String()
	resGetCoins, err := client.GetCoins(
		context.Background(),
		iotaclient.GetCoinsRequest{Owner: sentSigner.Address().AsIotaAddress(), CoinType: &coinType},
	)
	require.NoError(t, err)

	balancesSentSigner1, err := client.GetAllBalances(context.Background(), sentSigner.Address().AsIotaAddress())
	require.NoError(t, err)
	balancesStateSinger1, err := client.GetAllBalances(context.Background(), stateSigner.Address().AsIotaAddress())
	require.NoError(t, err)

	originDeposit := resGetCoins.Data[2]
	originDepositVal := coin.Value(originDeposit.Balance.Uint64())
	l1commitment := origin.L1Commitment(schemaVersion, initParams, iotago.ObjectID{}, originDepositVal, parameterstest.L1Mock)
	originStateMetadata := transaction.NewStateMetadata(
		schemaVersion,
		l1commitment,
		&iotago.ObjectID{},
		gas.DefaultFeePolicy(),
		initParams,
		originDepositVal,
		"https://iota.org",
	)
	gasCoin := resGetCoins.Data[0].Ref()
	txnResponse, anchorRef, err := startNewChain(
		t,
		client,
		&iscmoveclient.StartNewChainRequest{
			Signer:        sentSigner,
			AnchorOwner:   stateSigner.Address(),
			PackageID:     l1starter.ISCPackageID(),
			StateMetadata: originStateMetadata.Bytes(),
			InitCoinRef:   originDeposit.Ref(),
			GasPayments:   []*iotago.ObjectRef{gasCoin},
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	oriAnchor := isc.NewStateAnchor(anchorRef, l1starter.ISCPackageID())
	require.NotNil(t, oriAnchor)
	anchor, err := client.GetAnchorFromObjectID(context.Background(), oriAnchor.GetObjectID())
	require.NoError(t, err)
	require.EqualValues(t, oriAnchor.ChainID().AsAddress(), anchor.ObjectID)
	require.EqualValues(t, 0, anchor.Object.StateIndex)

	require.EqualValues(t, anchor.Object.StateMetadata, originStateMetadata.Bytes())

	balancesSentSinger2, err := client.GetAllBalances(context.Background(), sentSigner.Address().AsIotaAddress())
	require.NoError(t, err)
	require.EqualValues(t, balancesSentSigner1[0].TotalBalance.Int64()-originDeposit.Balance.Int64()-txnResponse.Effects.Data.GasFee(), balancesSentSinger2[0].TotalBalance.Int64())
	balancesStateSinger2, err := client.GetAllBalances(context.Background(), stateSigner.Address().AsIotaAddress())
	require.NoError(t, err)
	require.Equal(t, balancesStateSinger1[0], balancesStateSinger2[0])
}

// Was used to find proper deposit values for a specific metadata according to the existing hashes.
func TestMetadataBad(t *testing.T) {
	t.SkipNow()

	// This test was also skipped for wasp.
	// When it is enabled, it fails in both repos, so I have no easy way to decode that string of bytes
	// to be able then to re-implement it for isc.CallArguments instead of dict.Dict.

	// metadataHex := "0300000001006102000000e60701006204000000ffffffff01006322000000010024ed2ed9d3682c9c4b801dd15103f73d1fe877224cb51c8b3def6f91b67f5067"
	// metadataBin, err := hex.DecodeString(metadataHex)
	// require.NoError(t, err)
	// var initParams dict.Dict
	// initParams, err = dict.FromBytes(metadataBin)
	// require.NoError(t, err)
	// require.NotNil(t, initParams)
	// t.Logf("Args=%v", initParams)
	// initParams.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
	// 	t.Logf("Args, %v ===> %v", key, value)
	// 	return true
	// })

	// for deposit := uint64(0); deposit <= 10*isc.Million; deposit++ {
	// 	db := mapdb.NewMapDB()
	// 	st := state.NewStoreWithUniqueWriteMutex(db)
	// 	block1A, _ := origin.InitChain(0, st, initParams, deposit, isc.BaseTokenCoinInfo)
	// 	block1B, _ := origin.InitChain(0, st, initParams, 10*isc.Million-deposit, isc.BaseTokenCoinInfo)
	// 	block1C, _ := origin.InitChain(0, st, initParams, 10*isc.Million+deposit, isc.BaseTokenCoinInfo)
	// 	block2A, _ := origin.InitChain(0, st, nil, deposit, isc.BaseTokenCoinInfo)
	// 	block2B, _ := origin.InitChain(0, st, nil, 10*isc.Million-deposit, isc.BaseTokenCoinInfo)
	// 	block2C, _ := origin.InitChain(0, st, nil, 10*isc.Million+deposit, isc.BaseTokenCoinInfo)
	// 	t.Logf("Block0, deposit=%v => %v %v %v / %v %v %v", deposit,
	// 		block1A.L1Commitment(), block1B.L1Commitment(), block1C.L1Commitment(),
	// 		block2A.L1Commitment(), block2B.L1Commitment(), block2C.L1Commitment(),
	// 	)
	// }
}

func TestDictBytes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.SliceOfBytesMatching(".*").Draw(t, "key")
		val := rapid.SliceOfBytesMatching(".+").Draw(t, "val")
		d := dict.New()
		d.Set(kv.Key(key), val)
		b := d.Bytes()
		d2, err := dict.FromBytes(b)
		require.NoError(t, err)
		require.Equal(t, d, d2)
	})
}

// example values taken from a test on the testnet
func TestMismatchOriginCommitment(t *testing.T) {
	t.Skip("TODO")
	// store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	// oid, err := iotago.OutputIDFromHex("0xcf72dd6a8c8cd76eab93c80ae192677a17c554b91334a41bed5079eff37effc40000")
	// require.NoError(t, err)
	// originMetadata, err := cryptolib.DecodeHex("0x03016102e607016204ffffffff016322010024ed2ed9d3682c9c4b801dd15103f73d1fe877224cb51c8b3def6f91b67f5067")
	// require.NoError(t, err)
	// aoStateMetadata, err := cryptolib.DecodeHex("0x01000000006e55672af085d73ea0ed646f280a26e0eba053df10f439378fe4e99e0fb8774600761da7c0402da8640000000100000000010000000100000000")
	// require.NoError(t, err)
	// _, sender, err := iotago.ParseBech32("rms1qqjw6tke6d5ze8ztsqwaz5gr7u73l6rhyfxt28yt8hhklydk0agxwgerk65")
	// require.NoError(t, err)
	// _, stateController, err := iotago.ParseBech32("rms1qrkrlggl2plwfvxyuuyj55gw48ws0xwtteydez8y8e03elm3xf38gf7eq5r")
	// require.NoError(t, err)
	// _, govController, err := iotago.ParseBech32("rms1qqjw6tke6d5ze8ztsqwaz5gr7u73l6rhyfxt28yt8hhklydk0agxwgerk65")
	// require.NoError(t, err)
	// _, chainAliasAddress, err := iotago.ParseBech32("rms1pr27d4mr9wgesv8je5j6zkequhw0ysx55ftxt04z55dm9hc9yxkauqtukfl")
	// require.NoError(t, err)

	// ao := isc.NewAliasOutputWithID(
	// 	&iotago.AliasOutput{
	// 		Amount:         10000000,
	// 		NativeTokens:   []*iotago.NativeToken{},
	// 		AliasID:        chainAliasAddress.(*iotago.AliasAddress).AliasID(),
	// 		StateIndex:     0,
	// 		StateMetadata:  aoStateMetadata,
	// 		FoundryCounter: 0,
	// 		Conditions: []iotago.UnlockCondition{
	// 			&iotago.StateControllerAddressUnlockCondition{Address: stateController},
	// 			&iotago.GovernorAddressUnlockCondition{Address: govController},
	// 		},
	// 		Features: []iotago.Feature{
	// 			&iotago.SenderFeature{
	// 				Address: sender,
	// 			},
	// 			&iotago.MetadataFeature{Data: originMetadata},
	// 		},
	// 	},
	// 	oid,
	// )

	// _, err = origin.InitChainByAliasOutput(store, ao)
	// testmisc.RequireErrorToBe(t, err, "l1Commitment mismatch between originAO / originBlock")
}

func startNewChain(
	t *testing.T,
	client *iscmoveclient.Client,
	req *iscmoveclient.StartNewChainRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, *iscmove.RefWithObject[iscmove.Anchor], error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argInitCoin iotago.Argument
	if req.InitCoinRef != nil {
		ptb = iscmoveclient.PTBOptionSomeIotaCoin(ptb, req.InitCoinRef)
	} else {
		ptb = iscmoveclient.PTBOptionNoneIotaCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = iscmoveclient.PTBStartNewChain(ptb, req.PackageID, req.StateMetadata, argInitCoin, req.AnchorOwner)

	txnResponse, err := client.SignAndExecutePTB(
		context.Background(),
		req.Signer,
		ptb.Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
	require.NoError(t, err)

	anchorRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	require.NoError(t, err)
	anchor, err := client.GetAnchorFromObjectID(context.Background(), anchorRef.ObjectID)
	return txnResponse, anchor, err
}
