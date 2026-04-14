package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// executed in cluster_test.go
// deposit, withdraw, transfer
func (e *ChainEnv) testOffLedgerDepositWithdrawTransfer(t *testing.T) {
	keyPairUser1, addressUser1, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	_, addressUser2, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	userClient1 := e.NewChainClient(keyPairUser1)
	e.DepositFunds(10*isc.Million, keyPairUser1)
	balance1 := e.GetL1Balance(lo.ToPtr(addressUser1.AsIotaAddress()), coin.BaseTokenType)

	_, err = userClient1.PostOffLedgerRequest(context.Background(),
		accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
		},
	)
	require.NoError(t, err)
	time.Sleep(3 * time.Second)

	balance3 := e.GetL1Balance(lo.ToPtr(addressUser1.AsIotaAddress()), coin.BaseTokenType)
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
