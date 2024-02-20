package accounts_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

func TestAccounts(t *testing.T) {
	// execute tests on all schema versions
	for v := isc.SchemaVersion(0); v <= allmigrations.DefaultScheme.LatestSchemaVersion(); v++ {
		testCreditDebit1(t, v)
		testCreditDebit2(t, v)
		testCreditDebit3(t, v)
		testCreditDebit4(t, v)
		testCreditDebit5(t, v)
		testCreditDebit6(t, v)
		testCreditDebit7(t, v)
		testMoveAll(t, v)
		testDebitAll(t, v)
		testTransferNFTs(t, v)
		testCreditDebitNFT1(t, v)
	}
}

func knownAgentID(b byte, h uint32) isc.AgentID {
	var chainID isc.ChainID
	for i := range chainID {
		chainID[i] = b
	}
	return isc.NewContractAgentID(chainID, isc.Hname(h))
}

var dummyAssetID = [iotago.NativeTokenIDLength]byte{1, 2, 3}

func checkLedgerT(t *testing.T, v isc.SchemaVersion, state dict.Dict, cp string) *isc.Assets {
	require.NotPanics(t, func() {
		solo.CheckLedger(v, state, cp)
	})
	return accounts.GetTotalL2FungibleTokens(v, state)
}

func testCreditDebit1(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := knownAgentID(1, 2)
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, total.Equals(transfer))

	transfer.BaseTokens = 1
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp2")

	expected := isc.NewAssets(43, nil).AddNativeTokens(dummyAssetID, big.NewInt(4))
	require.True(t, expected.Equals(total))

	userAssets := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.EqualValues(t, 43, userAssets.BaseTokens)
	require.Zero(t, userAssets.NativeTokens.MustSet()[dummyAssetID].Amount.Cmp(big.NewInt(4)))
	checkLedgerT(t, v, state, "cp2")

	accounts.DebitFromAccount(v, state, agentID1, expected, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp3")
	expected = isc.NewEmptyAssets()
	require.True(t, expected.Equals(total))
}

func testCreditDebit2(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.DebitFromAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp2")
	require.EqualValues(t, 0, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil)
	require.True(t, expected.Equals(total))

	require.True(t, util.IsZeroBigInt(accounts.GetNativeTokenBalance(state, agentID1, transfer.NativeTokens[0].ID, isc.ChainID{})))
	bal1 := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.False(t, bal1.IsEmpty())
	require.True(t, total.Equals(bal1))
}

func testCreditDebit3(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(100))
	require.Panics(t,
		func() {
			accounts.DebitFromAccount(v, state, agentID1, transfer, isc.ChainID{})
		},
	)
	total = checkLedgerT(t, v, state, "cp2")

	require.EqualValues(t, 1, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))
}

func testCreditDebit4(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := accounts.AllAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokens(20)
	accounts.MustMoveBetweenAccounts(v, state, agentID1, agentID2, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp2")

	keys = accounts.AllAccountsAsDict(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.False(t, bm1.IsEmpty())
	expected = isc.NewAssets(22, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(bm1))

	bm2 := accounts.GetAccountFungibleTokens(v, state, agentID2, isc.ChainID{})
	require.False(t, bm2.IsEmpty())
	expected = isc.NewAssets(20, nil)
	require.True(t, expected.Equals(bm2))
}

func testCreditDebit5(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := accounts.AllAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokens(50)
	require.Error(t, accounts.MoveBetweenAccounts(v, state, agentID1, agentID2, transfer, isc.ChainID{}))
	total = checkLedgerT(t, v, state, "cp2")

	keys = accounts.AllAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.False(t, bm1.IsEmpty())
	require.True(t, expected.Equals(bm1))

	bm2 := accounts.GetAccountFungibleTokens(v, state, agentID2, isc.ChainID{})
	require.True(t, bm2.IsEmpty())
}

func testCreditDebit6(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	checkLedgerT(t, v, state, "cp1")

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	accounts.MustMoveBetweenAccounts(v, state, agentID1, agentID2, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state, "cp2")

	keys := accounts.AllAccountsAsDict(state).Keys()
	require.EqualValues(t, 2, len(keys))

	bal := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.True(t, bal.IsEmpty())

	bal2 := accounts.GetAccountFungibleTokens(v, state, agentID2, isc.ChainID{})
	require.False(t, bal2.IsEmpty())
	require.True(t, total.Equals(bal2))
}

func testCreditDebit7(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	checkLedgerT(t, v, state, "cp1")

	debitTransfer := isc.NewAssets(1, nil)
	// debit must fail
	require.Panics(t, func() {
		accounts.DebitFromAccount(v, state, agentID1, debitTransfer, isc.ChainID{})
	})

	total = checkLedgerT(t, v, state, "cp1")
	require.True(t, transfer.Equals(total))
}

func testMoveAll(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()
	agentID2 := isc.NewRandomAgentID()

	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	require.EqualValues(t, 1, accounts.AllAccountsMapR(state).Len())
	accs := accounts.AllAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	accounts.MustMoveBetweenAccounts(v, state, agentID1, agentID2, transfer, isc.ChainID{})
	require.EqualValues(t, 2, accounts.AllAccountsMapR(state).Len())
	accs = accounts.AllAccountsAsDict(state)
	require.EqualValues(t, 2, len(accs))
	_, ok = accs[kv.Key(agentID2.Bytes())]
	require.True(t, ok)
}

func testDebitAll(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()

	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.CreditToAccount(v, state, agentID1, transfer, isc.ChainID{})
	require.EqualValues(t, 1, accounts.AllAccountsMapR(state).Len())
	accs := accounts.AllAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	accounts.DebitFromAccount(v, state, agentID1, transfer, isc.ChainID{})
	require.EqualValues(t, 1, accounts.AllAccountsMapR(state).Len())
	accs = accounts.AllAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	require.True(t, ok)

	assets := accounts.GetAccountFungibleTokens(v, state, agentID1, isc.ChainID{})
	require.True(t, assets.IsEmpty())

	assets = accounts.GetTotalL2FungibleTokens(v, state)
	require.True(t, assets.IsEmpty())
}

func testTransferNFTs(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state, "cp0")

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	NFT1 := &isc.NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	accounts.CreditNFTToAccount(state, agentID1, &iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*iotago.NativeToken{},
		NFTID:        NFT1.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: NFT1.Issuer},
			&iotago.MetadataFeature{Data: NFT1.Metadata},
		},
	}, isc.ChainID{})
	// nft is credited
	user1NFTs := accounts.GetAccountNFTs(state, agentID1)
	require.Len(t, user1NFTs, 1)
	require.Equal(t, user1NFTs[0], NFT1.ID)

	// nft data is saved (accounts.SaveNFTOutput must be called)
	accounts.SaveNFTOutput(state, &iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*iotago.NativeToken{},
		NFTID:        NFT1.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: NFT1.Issuer},
			&iotago.MetadataFeature{Data: NFT1.Metadata},
		},
	}, 0)

	nftData := accounts.GetNFTData(state, NFT1.ID)
	require.Equal(t, nftData.ID, NFT1.ID)
	require.Equal(t, nftData.Issuer, NFT1.Issuer)
	require.Equal(t, nftData.Metadata, NFT1.Metadata)

	agentID2 := isc.NewRandomAgentID()

	// cannot move an NFT that is not owned
	require.Error(t, accounts.MoveBetweenAccounts(v, state, agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(iotago.NFTID{111}), isc.ChainID{}))

	// moves successfully when the NFT is owned
	accounts.MustMoveBetweenAccounts(v, state, agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(NFT1.ID), isc.ChainID{})

	user1NFTs = accounts.GetAccountNFTs(state, agentID1)
	require.Len(t, user1NFTs, 0)
	user2NFTs := accounts.GetAccountNFTs(state, agentID2)
	require.Len(t, user2NFTs, 1)
	require.Equal(t, user2NFTs[0], NFT1.ID)

	// remove the NFT from the chain
	accounts.DebitNFTFromAccount(state, agentID2, NFT1.ID, isc.ChainID{})
	require.Panics(t, func() {
		accounts.GetNFTData(state, NFT1.ID)
	})
}

func testCreditDebitNFT1(t *testing.T, _ isc.SchemaVersion) {
	state := dict.New()

	agentID1 := knownAgentID(1, 2)
	nft := isc.NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	accounts.CreditNFTToAccount(state, agentID1, &iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*iotago.NativeToken{},
		NFTID:        nft.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: nft.Issuer},
			&iotago.MetadataFeature{Data: nft.Metadata},
		},
	}, isc.ChainID{})

	accNFTs := accounts.GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 1)
	require.Equal(t, accNFTs[0], nft.ID)

	accounts.DebitNFTFromAccount(state, agentID1, nft.ID, isc.ChainID{})

	accNFTs = accounts.GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 0)
}
