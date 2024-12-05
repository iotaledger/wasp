package test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func TestOffLedger(t *testing.T) {
	l1starter.SingleTest(t)
	env := solo.New(t, &solo.InitOptions{
		GasBurnLogEnabled: true,
	})
	chain := env.NewChain()

	// create a wallet with some base tokens on L1:
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	chain.DepositBaseTokensToL2(env.L1BaseTokens(userAddress), userWallet)

	req := isc.NewOffLedgerRequest(chain.ID(), accounts.FuncDeposit.Message(), 0, math.MaxUint64)
	altReq := isc.NewImpersonatedOffLedgerRequest(req.(*isc.OffLedgerRequestData)).
		WithSenderAddress(userWallet.Address())

	rec, err := common.EstimateGas(chain, altReq)

	require.NoError(t, err)
	require.NotNil(t, rec)
	require.Greater(t, rec.GasFeeCharged, uint64(0))
}
