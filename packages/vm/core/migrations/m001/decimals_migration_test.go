package m001_test

import (
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/m001"
)

func TestM001Migration(t *testing.T) {
	// skipping, no need to store the snapshot and run this test after the migration is applied
	t.SkipNow()
	env := solo.New(t)

	// the snapshot is from commit 3c83b34
	// created by running TestSaveSnapshot in packages/solo/solotest
	env.RestoreSnapshot(env.LoadSnapshot("snapshot.db"))

	ch := env.GetChainByName("chain1")

	require.EqualValues(t, 5, ch.LatestBlockIndex())

	// add the migration to test
	ch.AddMigration(m001.AccountDecimals)

	// call views in pre-migration state
	// originator owns 17994760 tokens in the snapshot
	expectedOriginatorBalance := uint64(17994760)
	res, err := ch.CallView(accounts.Contract.Name, accounts.ViewBalanceBaseToken.Name, accounts.ParamAgentID, ch.OriginatorAgentID)
	require.NoError(t, err)
	require.EqualValues(t, expectedOriginatorBalance, codec.MustDecodeUint64(res.Get(accounts.ParamBalance)))

	checkGasEstimationWorks := func() {
		_, callData := solo.EVMCallDataFromArtifacts(t, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))
		_, err = ch.EVM().EstimateGas(ethereum.CallMsg{
			Data: callData,
		}, nil)
		require.NoError(ch.Env.T, err)
	}
	// gas estimation works on the pre-migrated state
	checkGasEstimationWorks()

	// cause a VM run, which will run the migration
	wallet, walletAddr := env.NewKeyPairWithFunds()
	err = ch.DepositBaseTokensToL2(1*isc.Million, wallet)
	require.NoError(t, err)

	// originator owns 17994760 tokens in the snapshot
	require.EqualValues(t, expectedOriginatorBalance+ch.LastReceipt().GasFeeCharged, ch.L2Assets(ch.OriginatorAgentID).BaseTokens)
	require.EqualValues(t, 1*isc.Million-ch.LastReceipt().GasFeeCharged, ch.L2Assets(isc.NewAgentID(walletAddr)).BaseTokens)
	// commonAccount owns 3000 tokens in the snapshot
	require.EqualValues(t, 3000, ch.L2Assets(accounts.CommonAccount()).BaseTokens)

	// gas call estimation still works after the migration
	checkGasEstimationWorks()
}
