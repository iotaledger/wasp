package m001_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
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

	// cause a VM run, which will run the migration
	wallet, walletAddr := env.NewKeyPairWithFunds()
	err := ch.DepositBaseTokensToL2(1*isc.Million, wallet)
	require.NoError(t, err)

	// originator owns 17994760 tokens in the snapshot
	require.EqualValues(t, 17994760+ch.LastReceipt().GasFeeCharged, ch.L2Assets(ch.OriginatorAgentID).BaseTokens)
	require.EqualValues(t, 1*isc.Million-ch.LastReceipt().GasFeeCharged, ch.L2Assets(isc.NewAgentID(walletAddr)).BaseTokens)
	// commonAccount owns 3000 tokens in the snapshot
	require.EqualValues(t, 3000, ch.L2Assets(accounts.CommonAccount()).BaseTokens)
}
