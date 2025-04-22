package test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestOffLedger(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		GasBurnLogEnabled: true,
	})
	chain := env.NewChain()

	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	chain.DepositBaseTokensToL2(env.L1BaseTokens(userAddress)/10, userWallet)

	req := isc.NewOffLedgerRequest(chain.ID(), accounts.FuncDeposit.Message(), 0, math.MaxUint64)

	altReq := isc.NewImpersonatedOffLedgerRequest(req.(*isc.OffLedgerRequestDataEssence)).
		WithSenderAddress(userWallet.Address())

	require.NotNil(t, altReq.SenderAccount())
	require.Equal(t, altReq.SenderAccount().String(), userWallet.Address().String())

	res := chain.EstimateGas(altReq)
	require.NotNil(t, res.Receipt)
	require.Nil(t, res.Receipt.Error)
	require.Greater(t, res.Receipt.GasFeeCharged, uint64(0))
}
