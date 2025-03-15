// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestAccessNodesOnLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Run("cluster=10, N=4, req=100", func(t *testing.T) {
		const numRequests = 100
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=15, N=6, req=200", func(t *testing.T) {
		testutil.RunHeavy(t)
		const numRequests = 200
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})
}

// This is the value of the Gas used per deposit
// This should probably be a bit nicer, than a hardcoded const hidden in a test :)
const BaseTokensDepositFee = 100

func testAccessNodesOnLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int) {
	cmt := util.MakeRange(0, numValidatorNodes)
	e := setupNativeInccounterTest(t, clusterSize, cmt)
	client, _ := e.NewRandomChainClient()

	for i := 0; i < numRequests; i++ {
		_, err := client.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			GasBudget:   iotaclient.DefaultGasBudget,
			Allowance:   isc.NewAssets(iotaclient.DefaultGasBudget),
			L2GasBudget: iotaclient.DefaultGasBudget,
			Transfer:    isc.NewAssets(iotaclient.DefaultGasBudget),
		})
		require.NoError(t, err)
	}

	expectedBalance := (iotaclient.DefaultGasBudget - BaseTokensDepositFee) * numRequests

	waitUntil(t, e.balanceEquals(isc.NewAddressAgentID(client.KeyPair.Address()), expectedBalance), e.Clu.AllNodes(), 40*time.Second, "a required number of testAccessNodesOnLedger requests")
}

func TestAccessNodesOffLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=10,N=4,req=50", func(t *testing.T) {
		const waitFor = 30 * time.Second
		const numRequests = 50
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=30,N=20,req=8", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 300 * time.Second
		const numRequests = 8
		const numValidatorNodes = 20
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})
}

func testAccessNodesOffLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int, timeout ...time.Duration) {
	to := 60 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	cmt := util.MakeRange(0, numValidatorNodes-1)

	e := setupNativeInccounterTest(t, clusterSize, cmt)

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := e.Chain.Client(keyPair)
	coinType := iotajsonrpc.IotaCoinType.String()
	balance, err := accountsClient.L1Client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{
		CoinType: &coinType,
		Owner:    accountsClient.KeyPair.Address().AsIotaAddress(),
	})

	require.NoError(t, err)

	tx, err := accountsClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(coin.Value(balance.Data[0].Balance.Uint64()) - iotaclient.DefaultGasBudget),
		GasBudget: iotaclient.DefaultGasBudget,
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, tx, true, 30*time.Second)
	require.NoError(t, err)

	client := e.Chain.Client(keyPair)

	someRandomsAddress := isc.NewAddressAgentID(cryptolib.NewRandomAddress())

	for i := 0; i < numRequests; i++ {
		_, err := client.PostOffLedgerRequest(context.Background(), accounts.FuncTransferAllowanceTo.Message(someRandomsAddress), chainclient.PostRequestParams{
			GasBudget:   iotaclient.DefaultGasBudget,
			Allowance:   isc.NewAssets(iotaclient.DefaultGasBudget),
			L2GasBudget: iotaclient.DefaultGasBudget,
			Transfer:    isc.NewAssets(iotaclient.DefaultGasBudget),
		})
		require.NoError(t, err)
	}

	expectedBalance := iotaclient.DefaultGasBudget * numRequests

	waitUntil(t, e.balanceEquals(someRandomsAddress, expectedBalance), util.MakeRange(0, clusterSize-1), to, "requests counted")
}
