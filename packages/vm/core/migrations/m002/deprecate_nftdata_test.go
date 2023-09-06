package m002_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/m002"
)

func TestM002Migration(t *testing.T) {
	// skipping, no need to store the snapshot and run this test after the migration is applied
	t.SkipNow()

	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})

	// the snapshot is from commit 54d70ac
	// created by running TestSaveSnapshot in packages/solo/solotest
	env.LoadSnapshot("snapshot.db")

	ch := env.GetChainByName("chain1")

	require.EqualValues(t, 5, ch.LatestBlockIndex())

	// add the migration to test
	ch.AddMigration(m002.DeprecateNFTData)

	// cause a VM run, which will run the migration
	err := ch.DepositBaseTokensToL2(1000, ch.OriginatorPrivateKey)
	require.NoError(t, err)

	// in the snapshot the OriginatorAgentID owns 1 NFT
	assets := ch.L2Assets(ch.OriginatorAgentID)
	require.Len(t, assets.NFTs, 1)
	nftID := assets.NFTs[0]

	// can still query the data of the NFT
	ret, err := ch.CallView(accounts.Contract.Name, accounts.ViewNFTData.Name, dict.Dict{
		accounts.ParamNFTID: nftID[:],
	})
	require.NoError(t, err)
	nft, err := isc.NFTFromBytes(ret.Get(accounts.ParamNFTData))
	require.NoError(t, err)
	require.NotNil(t, nft.Owner)
	require.NotNil(t, nft.Issuer)
	require.NotNil(t, nft.Metadata)
}
