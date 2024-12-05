package test

import (
	"testing"
)

func TestOffLedger(t *testing.T) {
	t.Fail()
	// TODO: Fix estimateGas(chain) -> chain does not match the interface
	/*

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
			require.Greater(t, rec.GasFeeCharged, uint64(0))*/
}
