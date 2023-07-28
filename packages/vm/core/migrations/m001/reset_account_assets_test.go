package m001_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/m001"
)

func TestM001Migration(t *testing.T) {
	// skipping, no need to store the snapshot and run this test after the migration is applied
	t.SkipNow()

	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})

	// the snapshot is from commit 7ab50c4dfae7a2ee445eea88651e8ef6bea66592
	// created by running TestSaveSnapshot in packages/solo/solotest
	env.LoadSnapshot("snapshot.db")

	ch := env.GetChainByName("chain1")

	require.EqualValues(t, 5, ch.LatestBlockIndex())

	// add the migration to test
	ch.AddMigration(m001.ResetAccountAssets)

	// cause a VM run, which will run the migration
	err := ch.DepositBaseTokensToL2(1000, ch.OriginatorPrivateKey)
	require.NoError(t, err)

	// in the snapshot the OriginatorAgentID owns a foundry and an NFT
	assets := ch.L2Assets(ch.OriginatorAgentID)
	require.Empty(t, assets.NFTs)
	require.Empty(t, assets.NativeTokens)

	// create a foundry again, test that the migration is not run again
	sn2, nativeTokenID2, err := ch.NewFoundryParams(1000).CreateFoundry()
	require.NoError(t, err)
	err = ch.MintTokens(sn2, 1000, ch.OriginatorPrivateKey)
	require.NoError(t, err)

	// cause a VM run
	err = ch.DepositBaseTokensToL2(1000, ch.OriginatorPrivateKey)
	require.NoError(t, err)

	ch.AssertL2NativeTokens(ch.OriginatorAgentID, nativeTokenID2, 1000)
}
