// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

func TestAccessNodesOnLedger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	t.Run("cluster=15, N=6, req=200", func(t *testing.T) {
		const numRequests = 200
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})
}

// This is the value of the Gas used per deposit
// This should probably be a bit nicer, than a hardcoded const hidden in a test :)
const BaseTokensDepositFee = 100_000

func testAccessNodesOnLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int) {
	cmt := util.MakeRange(0, numValidatorNodes)
	e := setupClusterTest(t, clusterSize, cmt)
	client, _ := e.NewRandomChainClient()

	for i := 0; i < numRequests; i++ {
		_, err := client.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			GasBudget:   iotagraphql.DefaultGasBudget,
			Allowance:   isc.NewAssets(iotagraphql.DefaultGasBudget),
			L2GasBudget: iotagraphql.DefaultGasBudget,
			Transfer:    isc.NewAssets(iotagraphql.DefaultGasBudget),
		})
		require.NoError(t, err)
	}

	expectedBalance := (iotagraphql.DefaultGasBudget - BaseTokensDepositFee) * numRequests

	waitUntil(t, e.balanceEquals(isc.NewAddressAgentID(client.KeyPair.Address()), expectedBalance), e.Clu.AllNodes(), 240*time.Second, fmt.Sprintf("balance to be %d", expectedBalance))
}

func TestAccessNodesOffLedger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	t.Run("cluster=30,N=20,req=8", func(t *testing.T) {
		const waitFor = 300 * time.Second
		const numRequests = 8
		const numValidatorNodes = 20
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})
}

func testAccessNodesOffLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int, timeout ...time.Duration) {
	to := 90 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	cmt := util.MakeRange(0, numValidatorNodes-1)

	e := setupClusterTest(t, clusterSize, cmt)

	accountsClient, _ := e.NewRandomChainClient()

	coinType := iotagraphql.IotaCoinType.String()
	balance, err := accountsClient.L1Client.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{
		CoinType: &coinType,
		Owner:    accountsClient.KeyPair.Address().AsIotaAddress(),
	})

	require.NoError(t, err)

	tx, err := accountsClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(coin.Value(balance.Data[0].Balance.Uint64()) - iotagraphql.DefaultGasBudget),
		GasBudget: iotagraphql.DefaultGasBudget,
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, tx, true, 30*time.Second)
	require.NoError(t, err)

	someRandomsAddress := isc.NewEthereumAddressAgentID(common.MaxAddress)

	nonce, err := accountsClient.ISCNonce(context.Background())
	require.NoError(t, err)

	for i := range numRequests {
		_, err2 := accountsClient.PostOffLedgerRequest(context.Background(), accounts.FuncTransferAllowanceTo.Message(someRandomsAddress), chainclient.PostRequestParams{
			Allowance: isc.NewAssets(iotagraphql.DefaultGasBudget),
			GasBudget: iotagraphql.DefaultGasBudget,
			Nonce:     nonce + uint64(i),
		})
		require.NoError(t, err2)
	}

	expectedBalance := iotagraphql.DefaultGasBudget * numRequests

	waitUntil(t, e.balanceEquals(someRandomsAddress, expectedBalance), util.MakeRange(0, clusterSize-1), to, "requests counted")
}
