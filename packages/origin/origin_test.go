package origin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

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
	"github.com/iotaledger/wasp/packages/kvstore/mapdb"
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

func TestInitChainByStateMetadataBytes(t *testing.T) {
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := origin.DefaultInitParams(isctest.NewRandomAgentID()).Encode()
	originDepositVal := coin.Value(100)

	var stateMetadata *transaction.StateMetadata
	{
		store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, stateMetadata = origin.InitChain(schemaVersion, store, initParams, iotago.ObjectID{}, originDepositVal, parameterstest.L1Mock)
	}

	{
		store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		stateMetadataBytes := stateMetadata.Bytes()
		_, err := origin.InitChainByStateMetadataBytes(store, stateMetadataBytes, originDepositVal, parameterstest.L1Mock)
		require.NoError(t, err)
	}

	{
		// corrupted/different state metadata
		stateMetadata := *stateMetadata
		stateMetadata.L1Commitment = &state.L1Commitment{}
		stateMetadataBytes := stateMetadata.Bytes()

		store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByStateMetadataBytes(store, stateMetadataBytes, originDepositVal, parameterstest.L1Mock)
		require.ErrorContains(t, err, "L1Commitment mismatch")
	}
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

	anchorRef, err := txnResponse.GetCreatedObjectByName(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	require.NoError(t, err)
	anchor, err := client.GetAnchorFromObjectID(context.Background(), anchorRef.ObjectID)
	return txnResponse, anchor, err
}
