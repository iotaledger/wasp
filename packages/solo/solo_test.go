package solo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
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
	t.Skip("FIXME cant hanlde the unmarshal of dry run result")
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
		iotagraphql.CoinValue(iotagraphql.DefaultGasBudget),
		iotagraphql.IotaCoinType,
	)
	ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
		ptb,
		l1starter.ISCPackageID(),
		argAssetsBag,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: testcoinRef}),
		iotagraphql.CoinValue(122),
		iotagraphql.CoinType(testcoinType.String()),
	)
	msg := &iscmove.Message{
		Contract: uint32(isc.Hn("accounts")),
		Function: uint32(isc.Hn("deposit")),
	}
	allowance := iscmove.NewAssets(33)
	allowance.SetCoin(iotagraphql.MustCoinTypeFromString(testcoinType.String()), iotagraphql.CoinValue(10))
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
		2*iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	dryRunRes, err := ch.Env.L1Client().DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: txBytes,
	})
	require.NoError(t, err)
	require.True(t, dryRunRes1.Effects.IsSuccess())

	estimateGasL1, err := ch.EstimateOnLedgerRequest(&dryRunRes.DryRunTransactionBlock)
	require.NoError(t, err)
	require.Nil(t, estimateGasL1.Receipt.Error)
	require.Greater(t, estimateGasL1.Receipt.GasBurned, uint64(0))
	require.Greater(t, estimateGasL1.Receipt.GasFeeCharged, uint64(0))
}
