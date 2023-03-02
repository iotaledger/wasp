package accounts

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

func knownAgentID(b byte, h uint32) isc.AgentID {
	var chainID isc.ChainID
	for i := range chainID {
		chainID[i] = b
	}
	return isc.NewContractAgentID(chainID, isc.Hname(h))
}

func TestBasic(t *testing.T) {
	t.Logf("Name: %s", Contract.Name)
	t.Logf("Description: %s", Contract.Description)
	t.Logf("Program hash: %s", Contract.ProgramHash.String())
	t.Logf("Hname: %s", Contract.Hname())
}

var dummyAssetID = [iotago.NativeTokenIDLength]byte{1, 2, 3}

func checkLedgerT(t *testing.T, state dict.Dict, cp string) *isc.Assets {
	require.NotPanics(t, func() {
		CheckLedger(state, cp)
	})
	return GetTotalL2FungibleTokens(state)
}

func TestCreditDebit1(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := knownAgentID(1, 2)
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, total.Equals(transfer))

	transfer.BaseTokens = 1
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp2")

	expected := isc.NewAssets(43, nil).AddNativeTokens(dummyAssetID, big.NewInt(4))
	require.True(t, expected.Equals(total))

	userAssets := GetAccountFungibleTokens(state, agentID1)
	require.EqualValues(t, 43, userAssets.BaseTokens)
	require.Zero(t, userAssets.NativeTokens.MustSet()[dummyAssetID].Amount.Cmp(big.NewInt(4)))
	checkLedgerT(t, state, "cp2")

	DebitFromAccount(state, agentID1, expected)
	total = checkLedgerT(t, state, "cp3")
	expected = isc.NewEmptyAssets()
	require.True(t, expected.Equals(total))
}

func TestCreditDebit2(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	DebitFromAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp2")
	require.EqualValues(t, 0, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil)
	require.True(t, expected.Equals(total))

	require.True(t, util.IsZeroBigInt(GetNativeTokenBalance(state, agentID1, transfer.NativeTokens[0].ID)))
	bal1 := GetAccountFungibleTokens(state, agentID1)
	require.False(t, bal1.IsEmpty())
	require.True(t, total.Equals(bal1))
}

func TestCreditDebit3(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	transfer = isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(100))
	require.Panics(t,
		func() {
			DebitFromAccount(state, agentID1, transfer)
		},
	)
	total = checkLedgerT(t, state, "cp2")

	require.EqualValues(t, 1, len(total.NativeTokens))
	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))
}

func TestCreditDebit4(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := allAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokens(20)
	MustMoveBetweenAccounts(state, agentID1, agentID2, transfer)
	total = checkLedgerT(t, state, "cp2")

	keys = allAccountsAsDict(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := GetAccountFungibleTokens(state, agentID1)
	require.False(t, bm1.IsEmpty())
	expected = isc.NewAssets(22, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(bm1))

	bm2 := GetAccountFungibleTokens(state, agentID2)
	require.False(t, bm2.IsEmpty())
	expected = isc.NewAssets(20, nil)
	require.True(t, expected.Equals(bm2))
}

func TestCreditDebit5(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.NativeTokens))
	require.True(t, expected.Equals(total))

	keys := allAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = isc.NewAssetsBaseTokens(50)
	require.Error(t, MoveBetweenAccounts(state, agentID1, agentID2, transfer))
	total = checkLedgerT(t, state, "cp2")

	keys = allAccountsAsDict(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := GetAccountFungibleTokens(state, agentID1)
	require.False(t, bm1.IsEmpty())
	require.True(t, expected.Equals(bm1))

	bm2 := GetAccountFungibleTokens(state, agentID2)
	require.True(t, bm2.IsEmpty())
}

func TestCreditDebit6(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	checkLedgerT(t, state, "cp1")

	agentID2 := isc.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	MustMoveBetweenAccounts(state, agentID1, agentID2, transfer)
	total = checkLedgerT(t, state, "cp2")

	keys := allAccountsAsDict(state).Keys()
	require.EqualValues(t, 2, len(keys))

	bal := GetAccountFungibleTokens(state, agentID1)
	require.True(t, bal.IsEmpty())

	bal2 := GetAccountFungibleTokens(state, agentID2)
	require.False(t, bal2.IsEmpty())
	require.True(t, total.Equals(bal2))
}

func TestCreditDebit7(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	transfer := isc.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	checkLedgerT(t, state, "cp1")

	debitTransfer := isc.NewAssets(1, nil)
	// debit must fail
	require.Panics(t, func() {
		DebitFromAccount(state, agentID1, debitTransfer)
	})

	total = checkLedgerT(t, state, "cp1")
	require.True(t, transfer.Equals(total))
}

func TestMoveAll(t *testing.T) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()
	agentID2 := isc.NewRandomAgentID()

	transfer := isc.NewAssetsBaseTokens(42).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	require.EqualValues(t, 1, allAccountsMapR(state).MustLen())
	accs := allAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	MustMoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.EqualValues(t, 2, allAccountsMapR(state).MustLen())
	accs = allAccountsAsDict(state)
	require.EqualValues(t, 2, len(accs))
	_, ok = accs[kv.Key(agentID2.Bytes())]
	require.True(t, ok)
}

func TestDebitAll(t *testing.T) {
	state := dict.New()
	agentID1 := isc.NewRandomAgentID()

	transfer := isc.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	require.EqualValues(t, 1, allAccountsMapR(state).MustLen())
	accs := allAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	DebitFromAccount(state, agentID1, transfer)
	require.EqualValues(t, 1, allAccountsMapR(state).MustLen())
	accs = allAccountsAsDict(state)
	require.EqualValues(t, 1, len(accs))
	require.True(t, ok)

	assets := GetAccountFungibleTokens(state, agentID1)
	require.True(t, assets.IsEmpty())

	assets = GetTotalL2FungibleTokens(state)
	require.True(t, assets.IsEmpty())
}

func TestTransferNFTs(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")

	require.True(t, total.Equals(isc.NewEmptyAssets()))

	agentID1 := isc.NewRandomAgentID()
	NFT1 := &isc.NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	CreditNFTToAccount(state, agentID1, NFT1)
	// nft is credited
	user1NFTs := getAccountNFTs(state, agentID1)
	require.Len(t, user1NFTs, 1)
	require.Equal(t, user1NFTs[0], NFT1.ID)

	// nft data is saved
	nftData := MustGetNFTData(state, NFT1.ID)
	require.Equal(t, nftData.ID, NFT1.ID)
	require.Equal(t, nftData.Issuer, NFT1.Issuer)
	require.Equal(t, nftData.Metadata, NFT1.Metadata)

	agentID2 := isc.NewRandomAgentID()

	// cannot move an NFT that is not owned
	require.Error(t, MoveBetweenAccounts(state, agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(iotago.NFTID{111})))

	// moves successfully when the NFT is owned
	MustMoveBetweenAccounts(state, agentID1, agentID2, isc.NewEmptyAssets().AddNFTs(NFT1.ID))

	user1NFTs = getAccountNFTs(state, agentID1)
	require.Len(t, user1NFTs, 0)
	user2NFTs := getAccountNFTs(state, agentID2)
	require.Len(t, user2NFTs, 1)
	require.Equal(t, user2NFTs[0], NFT1.ID)

	// remove the NFT from the chain
	DebitNFTFromAccount(state, agentID2, NFT1.ID)
	require.Panics(t, func() {
		MustGetNFTData(state, NFT1.ID)
	})
}

func TestFoundryOutputRec(t *testing.T) {
	o := foundryOutputRec{
		Amount: 300,
		TokenScheme: &iotago.SimpleTokenScheme{
			MaximumSupply: big.NewInt(1000),
			MintedTokens:  big.NewInt(20),
			MeltedTokens:  util.Big0,
		},
		BlockIndex:  3,
		OutputIndex: 2,
	}
	oBin := o.Bytes()
	o1, err := foundryOutputRecFromMarshalUtil(marshalutil.New(oBin))
	require.NoError(t, err)
	require.EqualValues(t, o.Amount, o1.Amount)
	ts, ok := o1.TokenScheme.(*iotago.SimpleTokenScheme)
	require.True(t, ok)
	//nolint:gocritic
	require.True(t, ts.MaximumSupply.Cmp(ts.MaximumSupply) == 0)
	//nolint:gocritic
	require.True(t, ts.MintedTokens.Cmp(ts.MintedTokens) == 0)
	require.EqualValues(t, o.BlockIndex, o1.BlockIndex)
	require.EqualValues(t, o.OutputIndex, o1.OutputIndex)
}

func TestCreditDebitNFT1(t *testing.T) {
	state := dict.New()

	agentID1 := knownAgentID(1, 2)
	nft := isc.NFT{
		ID:       iotago.NFTID{123},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: []byte("foobar"),
	}
	CreditNFTToAccount(state, agentID1, &nft)

	accNFTs := GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 1)
	require.Equal(t, accNFTs[0], nft.ID)

	DebitNFTFromAccount(state, agentID1, nft.ID)

	accNFTs = GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 0)
}
