package test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func TestOffLedger(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		GasBurnLogEnabled:        true,
	})
	chain := env.NewChain()

	// create a wallet with some base tokens on L1:
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	// the wallet can we identified on L2 by an AgentID:
	userAgentID := isc.NewAgentID(userAddress)
	// for now our on-chain account is empty:
	chain.AssertL2BaseTokens(userAgentID, 0)

	req := isc.NewOffLedgerRequest(chain.ID(), accounts.Contract.Hname(), accounts.ViewTotalAssets.Hname(), nil, 0, math.MaxUint64)
	req.WithSenderAddress(userWallet.Address())

	r := req.(isc.OffLedgerRequest)

	rec, err := common.EstimateGas(chain, r)

	require.NoError(t, err)
	require.NotNil(t, rec)
	require.Greater(t, rec.GasFeeCharged, uint64(0))
}
