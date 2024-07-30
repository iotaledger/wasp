package accounts_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
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

var dummyAssetID = [isc.NativeTokenIDLength]byte{1, 2, 3}

func checkLedgerT(t *testing.T, v isc.SchemaVersion, state dict.Dict) *isc.Assets {
	require.NotPanics(t, func() {
		accounts.NewStateReader(v, state).CheckLedgerConsistency()
	})
	return accounts.NewStateReader(v, state).GetTotalL2FungibleTokens()
}

func testCreditDebit1(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := knownAgentID(1, 2)
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	require.NotNil(t, total)
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, total.Equals(transfer))

	transfer.BaseTokens = 1
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	expected := isc.NewAssets(43, nil).AddNativeTokens(dummyAssetID, big.NewInt(4))
	require.True(t, expected.Equals(total))

	userAssets := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.EqualValues(t, 43, userAssets.BaseTokens)
	require.Zero(t, userAssets.NativeTokens.MustSet()[dummyAssetID].Amount.Cmp(big.NewInt(4)))
	checkLedgerT(t, v, state)

	accounts.NewStateWriter(v, state).DebitFromAccount(agentID1, expected, isc.ChainID{})
	total = checkLedgerT(t, v, state)
	expected = isc.NewEmptyAssets()
	require.True(t, expected.Equals(total))
}

func testCreditDebit2(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).DebitFromAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)
	require.EqualValues(t, 0, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil)
	require.True(t, expected.Equals(total))

	require.True(t, util.IsZeroBigInt(accounts.NewStateReader(v, state).GetNativeTokenBalance(agentID1, transfer.NativeTokens[0].ID, isc.ChainID{})))
	bal1 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.False(t, bal1.IsEmpty())
	require.True(t, total.Equals(bal1))
}

func testCreditDebit3(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(100))
	require.Panics(t,
		func() {
			accounts.NewStateWriter(v, state).DebitFromAccount(agentID1, transfer, isc.ChainID{})
		},
	)
	total = checkLedgerT(t, v, state)

	require.EqualValues(t, 1, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))
}

func testCreditDebit4(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokensU64(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := accounts.NewStateReader(v, state).AllAccountsAsDict().Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokensU64(20)
	err := accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, transfer, isc.ChainID{})
	require.NoError(t, err)
	total = checkLedgerT(t, v, state)

	keys = accounts.NewStateReader(v, state).AllAccountsAsDict().Keys()
	require.EqualValues(t, 2, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.False(t, bm1.IsEmpty())
	expected = isc.NewAssets(22, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(bm1))

	bm2 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID2, isc.ChainID{})
	require.False(t, bm2.IsEmpty())
	expected = isc.NewAssets(20, nil)
	require.True(t, expected.Equals(bm2))
}

func testCreditDebit5(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokensU64(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	total = checkLedgerT(t, v, state)

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := accounts.NewStateReader(v, state).AllAccountsAsDict().Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokensU64(50)
	require.Error(t, accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, transfer, isc.ChainID{}))
	total = checkLedgerT(t, v, state)

	keys = accounts.NewStateReader(v, state).AllAccountsAsDict().Keys()
	require.EqualValues(t, 1, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.False(t, bm1.IsEmpty())
	require.True(t, expected.Equals(bm1))

	bm2 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID2, isc.ChainID{})
	require.True(t, bm2.IsEmpty())
}

func testCreditDebit6(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokensU64(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	checkLedgerT(t, v, state)

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	err := accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, transfer, isc.ChainID{})
	require.NoError(t, err)
	total = checkLedgerT(t, v, state)

	keys := accounts.NewStateReader(v, state).AllAccountsAsDict().Keys()
	require.EqualValues(t, 2, len(keys))

	bal := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.True(t, bal.IsEmpty())

	bal2 := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID2, isc.ChainID{})
	require.False(t, bal2.IsEmpty())
	require.True(t, total.Equals(bal2))
}

func testCreditDebit7(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	checkLedgerT(t, v, state)

	debitTransfer := isc.NewAssets(1, nil)
	// debit must fail
	require.Panics(t, func() {
		accounts.NewStateWriter(v, state).DebitFromAccount(agentID1, debitTransfer, isc.ChainID{})
	})

	total = checkLedgerT(t, v, state)
	require.True(t, transfer.Equals(total))
}

func testMoveAll(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()
	agentID2 := isc.NewRandomAgentID()

	transfer := isc.NewAssetsBaseTokensU64(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	accs := accounts.NewStateReader(v, state).AllAccountsAsDict()
	require.Len(t, accs, 1)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	err := accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, transfer, isc.ChainID{})
	require.NoError(t, err)
	accs = accounts.NewStateReader(v, state).AllAccountsAsDict()
	require.Len(t, accs, 2)
	require.EqualValues(t, 2, len(accs))
	_, ok = accs[kv.Key(agentID2.Bytes())]
	require.True(t, ok)
}

func testDebitAll(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()

	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	accounts.NewStateWriter(v, state).CreditToAccount(agentID1, transfer, isc.ChainID{})
	accs := accounts.NewStateReader(v, state).AllAccountsAsDict()
	require.Len(t, accs, 1)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	accounts.NewStateWriter(v, state).DebitFromAccount(agentID1, transfer, isc.ChainID{})
	accs = accounts.NewStateReader(v, state).AllAccountsAsDict()
	require.Len(t, accs, 1)
	require.EqualValues(t, 1, len(accs))
	require.True(t, ok)

	assets := accounts.NewStateReader(v, state).GetAccountFungibleTokens(agentID1, isc.ChainID{})
	require.True(t, assets.IsEmpty())

	assets = accounts.NewStateReader(v, state).GetTotalL2FungibleTokens()
	require.True(t, assets.IsEmpty())
}

func testTransferNFTs(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()
	total := checkLedgerT(t, v, state)

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	NFT1 := &isc.NFT{
		ID:       isc.NFTID{123},
		Issuer:   cryptolib.NewRandomAddress(),
		Metadata: []byte("foobar"),
	}
	accounts.NewStateWriter(v, state).CreditNFTToAccount(agentID1, &iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*isc.NativeToken{},
		NFTID:        NFT1.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: NFT1.Issuer.AsIotagoAddress()},
			&iotago.MetadataFeature{Data: NFT1.Metadata},
		},
	}, isc.ChainID{})
	// nft is credited
	user1NFTs := accounts.NewStateReader(v, state).GetAccountNFTs(agentID1)
	require.Len(t, user1NFTs, 1)
	require.Equal(t, user1NFTs[0], NFT1.ID)

	// nft data is saved (accounts.SaveNFTOutput must be called)
	accounts.NewStateWriter(v, state).SaveNFTOutput(&iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*isc.NativeToken{},
		NFTID:        NFT1.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: NFT1.Issuer.AsIotagoAddress()},
			&iotago.MetadataFeature{Data: NFT1.Metadata},
		},
	}, 0)

	nftData := accounts.NewStateReader(v, state).GetNFTData(NFT1.ID)
	require.Equal(t, nftData.ID, NFT1.ID)
	require.Equal(t, nftData.Issuer, NFT1.Issuer)
	require.Equal(t, nftData.Metadata, NFT1.Metadata)

	agentID2 := isc.NewRandomAgentID()

	// cannot move an NFT that is not owned
	require.Error(t, accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(isc.NFTID{111}), isc.ChainID{}))

	// moves successfully when the NFT is owned
	err := accounts.NewStateWriter(v, state).MoveBetweenAccounts(agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(NFT1.ID), isc.ChainID{})
	require.NoError(t, err)

	user1NFTs = accounts.NewStateReader(v, state).GetAccountNFTs(agentID1)
	require.Len(t, user1NFTs, 0)
	user2NFTs := accounts.NewStateReader(v, state).GetAccountNFTs(agentID2)
	require.Len(t, user2NFTs, 1)
	require.Equal(t, user2NFTs[0], NFT1.ID)

	// remove the NFT from the chain
	accounts.NewStateWriter(v, state).DebitNFTFromAccount(agentID2, NFT1.ID, isc.ChainID{})
	require.Panics(t, func() {
		accounts.NewStateReader(v, state).GetNFTData(NFT1.ID)
	})
}

func testCreditDebitNFT1(t *testing.T, v isc.SchemaVersion) {
	state := dict.New()

	agentID1 := knownAgentID(1, 2)
	nft := isc.NFT{
		ID:       isc.NFTID{123},
		Issuer:   cryptolib.NewRandomAddress(),
		Metadata: []byte("foobar"),
	}
	accounts.NewStateWriter(v, state).CreditNFTToAccount(agentID1, &iotago.NFTOutput{
		Amount:       0,
		NativeTokens: []*isc.NativeToken{},
		NFTID:        nft.ID,
		ImmutableFeatures: []iotago.Feature{
			&iotago.IssuerFeature{Address: nft.Issuer.AsIotagoAddress()},
			&iotago.MetadataFeature{Data: nft.Metadata},
		},
	}, isc.ChainID{})

	accNFTs := accounts.NewStateReader(v, state).GetAccountNFTs(agentID1)
	require.Len(t, accNFTs, 1)
	require.Equal(t, accNFTs[0], nft.ID)

	accounts.NewStateWriter(v, state).DebitNFTFromAccount(agentID1, nft.ID, isc.ChainID{})

	accNFTs = accounts.NewStateReader(v, state).GetAccountNFTs(agentID1)
	require.Len(t, accNFTs, 0)
}
