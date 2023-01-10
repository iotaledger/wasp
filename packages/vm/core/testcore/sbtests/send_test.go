package sbtests

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestTooManyOutputsInASingleCall(t *testing.T) { run2(t, testTooManyOutputsInASingleCall) }
func testTooManyOutputsInASingleCall(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	// send 1 tx will 1_000_000 BaseTokens which should result in too many outputs, so the request must fail
	wallet, _ := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddBaseTokens(1000 * isc.Million).
		AddAllowanceBaseTokens(999 * isc.Million). // contract is sending 1Mi per output
		WithGasBudget(math.MaxUint64)
	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSeveralOutputsInASingleCall(t *testing.T) { run2(t, testSeveralOutputsInASingleCall) }
func testSeveralOutputsInASingleCall(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, walletAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositBaseTokensToL2(100_000, wallet)
	require.NoError(t, err)

	beforeWallet := ch.L1L2Funds(walletAddr)
	t.Logf("----- BEFORE wallet: %s", beforeWallet)

	// this will SUCCEED because it will result in 4 outputs in the single call
	const allowance = 4 * isc.Million
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddAllowanceBaseTokens(allowance).
		AddBaseTokens(allowance + 1*isc.Million).
		WithGasBudget(math.MaxUint64)
	tx, _, err := ch.PostRequestSyncTx(req, wallet)
	require.NoError(t, err)

	storageDeposit := tx.Essence.Outputs[0].Deposit()
	ch.Env.AssertL1BaseTokens(walletAddr, beforeWallet.AssetsL1.BaseTokens+allowance-storageDeposit)
}

func TestSeveralOutputsInASingleCallFail(t *testing.T) { run2(t, testSeveralOutputsInASingleCallFail) }
func testSeveralOutputsInASingleCallFail(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, walletAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositBaseTokensToL2(100_000, wallet)
	require.NoError(t, err)

	beforeWallet := ch.L1L2Funds(walletAddr)
	t.Logf("----- BEFORE wallet: %s", beforeWallet)

	// this will FAIL because it will result in 5 outputs in the single call
	const allowance = 5 * isc.Million
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddAllowanceBaseTokens(allowance).
		AddBaseTokens(allowance + 1*isc.Million).
		WithGasBudget(math.MaxUint64)

	_, err = ch.PostRequestSync(req, wallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSplitTokensFail(t *testing.T) { run2(t, testSplitTokensFail) }
func testSplitTokensFail(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, _ := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositBaseTokensToL2(2*isc.Million, wallet)
	require.NoError(t, err)

	sn, nativeTokenID, err := ch.NewFoundryParams(100).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	err = ch.MintTokens(sn, big.NewInt(100), wallet)
	require.NoError(t, err)

	// this will FAIL because it will result in 100 outputs in the single call
	allowance := isc.NewAllowanceBaseTokens(100*isc.Million).AddNativeTokens(nativeTokenID, 100)
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFundsNativeTokens.Name).
		AddAllowance(allowance).
		AddBaseTokens(200 * isc.Million).
		WithGasBudget(math.MaxUint64)
	_, err = ch.PostRequestSync(req, wallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSplitTokensSuccess(t *testing.T) { run2(t, testSplitTokensSuccess) }
func testSplitTokensSuccess(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, addr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))
	agentID := isc.NewAgentID(addr)

	err := ch.DepositBaseTokensToL2(2*isc.Million, wallet)
	require.NoError(t, err)

	amountMintedTokens := int64(100)
	sn, nativeTokenID, err := ch.NewFoundryParams(amountMintedTokens).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	err = ch.MintTokens(sn, big.NewInt(amountMintedTokens), wallet)
	require.NoError(t, err)

	amountTokensToSend := int64(3)
	allowance := isc.NewAllowanceBaseTokens(2*isc.Million).AddNativeTokens(nativeTokenID, amountTokensToSend)
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFundsNativeTokens.Name).
		AddAllowance(allowance).
		AddBaseTokens(2 * isc.Million).
		WithGasBudget(math.MaxUint64)
	_, err = ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
	require.Equal(t, ch.L2NativeTokens(agentID, nativeTokenID).Int64(), amountMintedTokens-amountTokensToSend)
	require.Equal(t, ch.Env.L1NativeTokens(addr, nativeTokenID).Int64(), amountTokensToSend)
}

func TestPingBaseTokens1(t *testing.T) { run2(t, testPingBaseTokens1) }
func testPingBaseTokens1(t *testing.T, w bool) {
	// TestPingBaseTokens1 sends some base tokens to SC and receives the whole allowance sent back to L1 as on-ledger request
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	user, userAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(1))

	userFundsBefore := ch.L1L2Funds(userAddr)
	commonBefore := ch.L2CommonAccountAssets()
	t.Logf("----- BEFORE -----\nUser funds left: %s\nCommon account: %s", userFundsBefore, commonBefore)

	const expectedBack = 1 * isc.Million
	ch.Env.AssertL1BaseTokens(userAddr, utxodb.FundsFromFaucetAmount)

	req := solo.NewCallParams(ScName, sbtestsc.FuncPingAllowanceBack.Name).
		AddBaseTokens(expectedBack + 10_000). // add extra base tokens besides allowance in order to estimate the gas fees
		AddAllowanceBaseTokens(expectedBack)

	gas, gasFee, err := ch.EstimateGasOnLedger(req, user, true)
	require.NoError(t, err)
	req.
		WithFungibleTokens(isc.NewFungibleBaseTokens(expectedBack + gasFee)).
		WithGasBudget(gas + 1)

	_, err = ch.PostRequestSync(req, user)
	require.NoError(t, err)
	receipt := ch.LastReceipt()

	userFundsAfter := ch.L1L2Funds(userAddr)
	commonAfter := ch.L2CommonAccountAssets()
	t.Logf("------ AFTER ------\nReceipt: %s\nUser funds left: %s\nCommon account: %s", receipt, userFundsAfter, commonAfter)

	require.EqualValues(t, userFundsAfter.AssetsL1.BaseTokens, utxodb.FundsFromFaucetAmount-receipt.GasFeeCharged)
	require.EqualValues(t, int(commonBefore.BaseTokens+receipt.GasFeeCharged), int(commonAfter.BaseTokens))
	require.EqualValues(t, utxodb.FundsFromFaucetAmount-receipt.GasFeeCharged, userFundsAfter.AssetsL1.BaseTokens)
	require.Zero(t, userFundsAfter.AssetsL2.BaseTokens)
}

func TestEstimateMinimumStorageDeposit(t *testing.T) { run2(t, testEstimateMinimumStorageDeposit) }
func testEstimateMinimumStorageDeposit(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, _ := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	// should fail without enough base tokens to pay for a L1 transaction storage deposit
	allowance := isc.NewAllowanceBaseTokens(1)
	req := solo.NewCallParams(ScName, sbtestsc.FuncEstimateMinStorageDeposit.Name).
		AddAllowance(allowance).
		AddBaseTokens(100_000).
		WithGasBudget(100_000)

	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err)

	// should succeed with enough base tokens to pay for a L1 transaction storage deposit
	allowance = isc.NewAllowanceBaseTokens(100_000)
	req = solo.NewCallParams(ScName, sbtestsc.FuncEstimateMinStorageDeposit.Name).
		AddAllowance(allowance).
		AddBaseTokens(100_000).
		WithGasBudget(100_000)

	_, err = ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
}

func mintDummyNFT(t *testing.T, ch *solo.Chain, issuer *cryptolib.KeyPair, owner iotago.Address) (*isc.NFT, *solo.NFTMintedInfo) {
	nftMetadata := []byte("foobar")
	nft, nftInfo, err := ch.Env.MintNFTL1(issuer, owner, nftMetadata)
	require.NoError(t, err)
	return nft, nftInfo
}

func TestSendNFTsBack(t *testing.T) { run2(t, testSendNFTsBack) }
func testSendNFTsBack(t *testing.T, w bool) {
	// Send NFT and receive it back (on-ledger request)
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, addr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(0))

	nft, _ := mintDummyNFT(t, ch, wallet, addr)

	baseTokensToSend := uint64(300_000)
	baseTokensForGas := uint64(100_000)
	assetsToSend := isc.NewFungibleBaseTokens(baseTokensToSend)
	assetsToAllow := isc.NewFungibleBaseTokens(baseTokensToSend - baseTokensForGas)

	// receive an NFT back that is sent in the same request
	req := solo.NewCallParams(ScName, sbtestsc.FuncSendNFTsBack.Name).
		AddFungibleTokens(assetsToSend).
		WithNFT(nft).
		AddAllowance(isc.NewAllowanceFungibleTokens(assetsToAllow).AddNFTs(nft.ID)).
		WithMaxAffordableGasBudget()

	_, err := ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
	require.True(t, ch.Env.HasL1NFT(addr, &nft.ID))
}

func TestNFTOffledgerWithdraw(t *testing.T) { run2(t, testNFTOffledgerWithdraw) }
func testNFTOffledgerWithdraw(t *testing.T, w bool) {
	// Deposit an NFT, then claim it back via offleger-request
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, issuerAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(0))

	nft, _ := mintDummyNFT(t, ch, wallet, issuerAddr)

	require.True(t, ch.Env.HasL1NFT(issuerAddr, &nft.ID))
	require.False(t, ch.Env.HasL1NFT(ch.ChainID.AsAddress(), &nft.ID))
	require.False(t, ch.HasL2NFT(isc.NewAgentID(issuerAddr), &nft.ID))

	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		AddFungibleTokens(isc.NewFungibleBaseTokens(1_000_000)).
		WithNFT(nft).
		WithMaxAffordableGasBudget()

	_, err := ch.PostRequestSync(req, wallet)
	require.NoError(t, err)

	require.False(t, ch.Env.HasL1NFT(issuerAddr, &nft.ID))
	require.True(t, ch.Env.HasL1NFT(ch.ChainID.AsAddress(), &nft.ID))
	require.True(t, ch.HasL2NFT(isc.NewAgentID(issuerAddr), &nft.ID))

	wdReq := solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		WithAllowance(isc.NewAllowance(10_000, nil, []iotago.NFTID{nft.ID})).
		WithMaxAffordableGasBudget()

	_, err = ch.PostRequestOffLedger(wdReq, wallet)
	require.NoError(t, err)

	require.True(t, ch.Env.HasL1NFT(issuerAddr, &nft.ID))
	require.False(t, ch.Env.HasL1NFT(ch.ChainID.AsAddress(), &nft.ID))
	require.False(t, ch.HasL2NFT(isc.NewAgentID(issuerAddr), &nft.ID))
}

func TestNFTMintToChain(t *testing.T) { run2(t, testNFTMintToChain) }
func testNFTMintToChain(t *testing.T, w bool) {
	// Mints an NFT as a request
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, addr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(0))

	nftToBeMinted := &isc.NFT{
		ID:       iotago.NFTID{},
		Issuer:   addr,
		Metadata: []byte("foobar"),
	}

	baseTokensToSend := uint64(300_000)
	baseTokensForGas := uint64(100_000)
	assetsToSend := isc.NewFungibleBaseTokens(baseTokensToSend)
	assetsToAllow := isc.NewFungibleBaseTokens(baseTokensToSend - baseTokensForGas)

	// receive an NFT back that is sent in the same request
	req := solo.NewCallParams(ScName, sbtestsc.FuncClaimAllowance.Name).
		AddFungibleTokens(assetsToSend).
		WithNFT(nftToBeMinted).
		AddAllowance(isc.NewAllowanceFungibleTokens(assetsToAllow).AddNFTs(iotago.NFTID{})). // empty NFTID
		WithMaxAffordableGasBudget()

	_, err := ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
	// find out the NFTID
	receipt := ch.LastReceipt()
	nftID := iotago.NFTIDFromOutputID(receipt.DeserializedRequest().ID().OutputID())

	// - Chain owns the NFT on L1
	require.True(t, ch.Env.HasL1NFT(ch.ChainID.AsAddress(), &nftID))
	// - The target contract owns the NFT on L2
	contractAgentID := isc.NewContractAgentID(ch.ChainID, sbtestsc.Contract.Hname())
	require.True(t, ch.HasL2NFT(contractAgentID, &nftID))
}
