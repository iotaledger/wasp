package solo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestSoloBasic1(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain(false)
	require.Zero(env.T, ch.L2CommonAccountAssets().Coins.BaseTokens())
	require.Zero(env.T, ch.L2BaseTokens(ch.AdminAgentID()))

	err := ch.DepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)
	require.NoError(env.T, err)
	require.NotZero(env.T, ch.L2BaseTokens(ch.AdminAgentID()))
}

func TestDryRunForRequest(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain(false)
	sender := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	coinPackageID, treasuryCap := ch.Env.L1DeployCoinPackage(sender)
	testcoinType := coin.MustTypeFromString(fmt.Sprintf(
		"%s::%s::%s",
		coinPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	))
	testcoinRef := ch.Env.L1MintCoin(
		sender,
		coinPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		1*isc.Million,
	)

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = iscmoveclient.PTBAssetsBagNew(ptb, l1starter.ISCPackageID(), sender.Address())
	argAssetsBag := ptb.LastCommandResultArg()
	ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
		ptb,
		l1starter.ISCPackageID(),
		argAssetsBag,
		iotago.GetArgumentGasCoin(),
		iotajsonrpc.CoinValue(iotaclient.DefaultGasBudget),
		iotajsonrpc.IotaCoinType,
	)
	ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
		ptb,
		l1starter.ISCPackageID(),
		argAssetsBag,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: testcoinRef}),
		iotajsonrpc.CoinValue(122),
		iotajsonrpc.CoinType(testcoinType.String()),
	)
	msg := &iscmove.Message{
		Contract: uint32(isc.Hn("accounts")),
		Function: uint32(isc.Hn("deposit")),
	}
	allowance := iscmove.NewAssets(33)
	allowance.AddCoin(iotajsonrpc.MustCoinTypeFromString(testcoinType.String()), iotajsonrpc.CoinValue(10))
	req := iscmoveclient.PTBCreateAndSendRequest(
		ptb,
		l1starter.ISCPackageID(),
		ch.ChainID.AsObjectID(),
		argAssetsBag,
		msg, bcs.MustMarshal(allowance), 1074)

	tx := req.Finish()

	txData := iotago.NewProgrammable(
		sender.Address().AsIotaAddress(),
		tx,
		[]*iotago.ObjectRef{},
		2*iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	dryRunRes1, err := ch.Env.L1Client().DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.True(t, dryRunRes1.Effects.Data.IsSuccess())

	var dryRunRes2 iotajsonrpc.DryRunTransactionBlockResponse
	b, err := bcs.Marshal(dryRunRes1)
	require.NoError(t, err)
	dryRunRes2, err = bcs.Unmarshal[iotajsonrpc.DryRunTransactionBlockResponse](b)
	require.NoError(t, err)
	estimateGasL1, err := ch.EstimateGasL1(&dryRunRes2)
	require.NoError(t, err)
	require.Nil(t, estimateGasL1.Receipt.Error)
	require.Greater(t, estimateGasL1.Receipt.GasBurned, uint64(0))
	require.Greater(t, estimateGasL1.Receipt.GasFeeCharged, uint64(0))
}
