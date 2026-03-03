package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// executed in cluster_test.go
func (e *ChainEnv) testOnLedgerDeposit(t *testing.T) {
	userWallet, userAddr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	userClient := e.Chain.Client(userWallet)
	balance1 := e.GetL2Balance(isc.NewAddressAgentID(userAddr), coin.BaseTokenType)

	tx := [5]*iotagraphql.ExecuteTransactionBlockResponse{}
	gasFeeChargedSum := coin.Value(0)
	baseTokesSent := coin.Value(10 + iotagraphql.DefaultGasBudget)
	for i := 0; i < 5; i++ {
		tx[i], err = userClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(baseTokesSent),
			GasBudget: iotagraphql.DefaultGasBudget,
		})
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, tx[i], false, 30*time.Second)
		require.NoError(t, err)

		gasFeeCharged, err := util.DecodeUint64(receipts[0].GasFeeCharged)
		require.NoError(t, err)

		gasFeeChargedSum += coin.Value(gasFeeCharged)
	}

	balance2 := e.GetL2Balance(isc.NewAddressAgentID(userAddr), coin.BaseTokenType)
	require.Equal(t, balance1+5*coin.Value(10+iotagraphql.DefaultGasBudget)-gasFeeChargedSum, balance2)
}
