package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

// executed in cluster_test.go
func (e *ChainEnv) testBasicAccounts(t *testing.T) {
	e.testAccounts()
}

func TestBasicAccountsNLow(t *testing.T) {
	runTest := func(tt *testing.T, n, t int) {
		clu := newCluster(tt)
		chainNodes := make([]int, n)
		for i := range chainNodes {
			chainNodes[i] = i
		}
		chain, err := clu.DeployChainWithDKG(chainNodes, chainNodes, uint16(t))
		require.NoError(tt, err)
		env := newChainEnv(tt, clu, chain)
		env.testAccounts()
	}
	t.Run("N=1", func(tt *testing.T) { runTest(tt, 1, 1) })
	t.Run("N=2", func(tt *testing.T) { runTest(tt, 2, 2) })
	t.Run("N=3", func(tt *testing.T) { runTest(tt, 3, 3) })
	t.Run("N=4", func(tt *testing.T) { runTest(tt, 4, 3) })
}

func (e *ChainEnv) testAccounts() {
	e.t.Logf("   %s: %s", root.Contract.Name, root.Contract.Hname().String())
	e.t.Logf("   %s: %s", accounts.Contract.Name, accounts.Contract.Hname().String())

	e.checkCoreContracts()

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	originatorClient := e.NewChainClient(keyPair)
	_, err = originatorClient.DepositFunds(1 * isc.Million)
	require.NoError(e.t, err)
	time.Sleep(3 * time.Second)
	balance1, err := originatorClient.L1Client.GetBalance(context.TODO(), iotaclient.GetBalanceRequest{Owner: keyPair.Address().AsIotaAddress()})
	require.NoError(e.t, err)

	balance2 := e.GetL1Balance(keyPair.Address().AsIotaAddress(), coin.BaseTokenType)
	require.Equal(e.t, balance1.TotalBalance.Uint64(), balance2.Uint64())

	_, err = originatorClient.PostOffLedgerRequest(context.Background(),
		accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
		},
	)
	require.NoError(e.t, err)
	time.Sleep(3 * time.Second)

	balance3 := e.GetL1Balance(keyPair.Address().AsIotaAddress(), coin.BaseTokenType)
	require.Equal(e.t, balance1.TotalBalance.Uint64()+10, balance3.Uint64())
}

// executed in cluster_test.go
func (e *ChainEnv) testBasic2Accounts(t *testing.T) {
	e.checkCoreContracts()

	keyPairUser1, addressUser1, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	_, addressUser2, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	userClient1 := e.NewChainClient(keyPairUser1)
	userClient1.DepositFunds(10 * isc.Million)
	time.Sleep(3 * time.Second)
	balance1 := e.GetL1Balance(addressUser1.AsIotaAddress(), coin.BaseTokenType)

	balance2 := e.GetL1Balance(addressUser1.AsIotaAddress(), coin.BaseTokenType)
	require.Equal(t, balance1, balance2)

	_, err = userClient1.PostOffLedgerRequest(context.Background(),
		accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
		},
	)
	require.NoError(t, err)
	time.Sleep(3 * time.Second)

	balance3 := e.GetL1Balance(addressUser1.AsIotaAddress(), coin.BaseTokenType)
	require.Equal(t, balance1+10, balance3)

	user1L2Bal1 := e.GetL2Balance(isc.NewAddressAgentID(addressUser1), coin.BaseTokenType)
	user2L2Bal1 := e.GetL2Balance(isc.NewAddressAgentID(addressUser2), coin.BaseTokenType)

	var transferAmount coin.Value = 10
	req, err := userClient1.PostOffLedgerRequest(context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(addressUser2)),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(transferAmount),
		},
	)
	require.NoError(t, err)
	time.Sleep(3 * time.Second)

	reqceipt, err := e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, req.ID(), false, 30*time.Second)
	require.NoError(t, err)

	user1L2Bal2 := e.GetL2Balance(isc.NewAddressAgentID(addressUser1), coin.BaseTokenType)
	require.NoError(t, err)
	user2L2Bal2 := e.GetL2Balance(isc.NewAddressAgentID(addressUser2), coin.BaseTokenType)
	require.NoError(t, err)
	gasFeeCharged, err := strconv.ParseUint(reqceipt.GasFeeCharged, 10, 64)
	require.NoError(t, err)
	require.Equal(t, user1L2Bal1-coin.Value(gasFeeCharged)-transferAmount, user1L2Bal2)
	require.Equal(t, user2L2Bal1+transferAmount, user2L2Bal2)
}
