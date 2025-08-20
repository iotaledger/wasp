package solo_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
)

// This test is an example of how to generate a snapshot from a Solo chain.
// The snapshot is especially useful to test migrations.
func TestSaveSnapshot(t *testing.T) {
	// skipped by default because the generated dump is fairly large
	if os.Getenv("ENABLE_SOLO_SNAPSHOT") == "" {
		t.SkipNow()
	}

	t.Fatalf("TODO")

	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()
	ch.MustDepositBaseTokensToL2(2*isc.Million, ch.ChainAdmin)

	// create foundry and native tokens on L2
	// sn, nativeTokenID, err := ch.NewNativeTokenParams(1000).CreateFoundry()
	// require.NoError(t, err)
	// // mint some tokens for the user
	// err = ch.MintTokens(sn, 1000, ch.OwnerPrivateKey)
	// require.NoError(t, err)

	// _, err = ch.GetNativeTokenIDByFoundrySN(sn)
	// require.NoError(t, err)
	// ch.AssertL2Coins(ch.OriginatorAgentID, nativeTokenID, 1000)

	// TODO impl NFT
	// create NFT on L1 and deposit on L2
	// nft, _, err := ch.Env.MintNFTL1(ch.OwnerPrivateKey, ch.OriginatorAddress, []byte("foobar"))
	// require.NoError(t, err)
	// _, err = ch.PostRequestSync(
	// 	solo.NewCallParams(accounts.FuncDeposit.Message()).
	// 		WithNFT(nft).
	// 		AddBaseTokens(10*isc.Million).
	// 		WithMaxAffordableGasBudget(),
	// 	ch.OwnerPrivateKey)
	// require.NoError(t, err)

	// require.NotEmpty(t, ch.L2NFTs(ch.OriginatorAgentID))

	ch.Env.SaveSnapshot(ch.Env.TakeSnapshot(), "snapshot.db")
}

// This test is an example of how to restore a Solo snapshot.
// The snapshot is especially useful to test migrations.
func TestLoadSnapshot(t *testing.T) {
	// skipped because this is just an example, the dump is not committed
	t.SkipNow()

	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	env.RestoreSnapshot(env.LoadSnapshot("snapshot.db"))

	ch := env.GetChainByName("chain1")

	require.EqualValues(t, 5, ch.LatestBlockIndex())

	// nativeTokenID, err := ch.GetNativeTokenIDByFoundrySN(1)
	// require.NoError(t, err)
	// ch.AssertL2Coins(ch.OriginatorAgentID, nativeTokenID, big.NewInt(1000))
}
