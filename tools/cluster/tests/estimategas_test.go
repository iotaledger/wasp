package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func testEstimateGasOnLedger(t *testing.T, env *ChainEnv) {
	// estimate on-ledger request, then send the same request, assert the gas used/fees match
	panic("refactor me: transaction.BasicOutputFromPostData")
	var output iotago.Output

	outputBytes, err := output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	require.NoError(t, err)

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOnledger{
		OutputBytes: cryptolib.EncodeHex(outputBytes),
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	feeCharged, err := strconv.ParseUint(estimatedReceipt.GasFeeCharged, 10, 64)
	require.NoError(t, err)

	client := env.Chain.Client(keyPair)
	par := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(coin.Value(feeCharged)),
		Allowance: isc.NewAssets(5000),
	}
	gasBudget, err := strconv.ParseUint(estimatedReceipt.GasBurned, 10, 64)
	require.NoError(t, err)
	par.WithGasBudget(gasBudget)

	tx, err := client.PostRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		par,
	)
	require.NoError(t, err)
	recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.GasFeeCharged)
}

func testEstimateGasOnLedgerNFT(t *testing.T, env *ChainEnv) {
	// estimate on-ledger request, using and NFT with minSD
	t.Fail()
	/*
		keyPair, addr, err := env.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)

		metadata, err := cryptolib.DecodeHex("0x7b227374616e64617264223a224952433237222c2276657273696f6e223a2276312e30222c226e616d65223a2254657374416761696e4e667432222c2274797065223a22696d6167652f6a706567222c22757269223a2268747470733a2f2f696d616765732e756e73706c6173682e636f6d2f70686f746f2d313639353539373737383238392d6663316635633731353935383f69786c69623d72622d342e302e3326697869643d4d3377784d6a4133664442384d48787761473930627931775957646c664878386647567566444238664878386641253344253344266175746f3d666f726d6174266669743d63726f7026773d3335343226713d3830227d")
		require.NoError(t, err)

		nftID, _, err := env.Clu.MintL1NFT(metadata, addr, keyPair)
		require.NoError(t, err)
		nft := &isc.NFT{
			ID:       iotago.NFTIDFromOutputID(nftID),
			Issuer:   addr,
			Metadata: metadata,
		}

		targetAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, common.Address{})

		panic("refactor me: transaction.NFTOutputFromPostData")

		var output iotago.NFTOutput
		outputBytes, err := output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
		require.NoError(t, err)

		estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background(),
			env.Chain.ChainID.String(),
		).Request(apiclient.EstimateGasRequestOnledger{
			OutputBytes: cryptolib.EncodeHex(outputBytes),
		}).Execute()
		require.NoError(t, err)
		require.Empty(t, estimatedReceipt.ErrorMessage)

		client := env.Chain.Client(keyPair)
		par := chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(coin.Value(output.Deposit())),
			Allowance: isc.NewEmptyAssets().AddObject(nft.ID),
			NFT:       nft,
		}
		gasBudget, err := strconv.ParseUint(estimatedReceipt.GasBurned, 10, 64)
		require.NoError(t, err)
		par.WithGasBudget(gasBudget)

		tx, err := client.PostRequest(context.Background(), accounts.FuncTransferAllowanceTo.Message(targetAgentID), par)
		require.NoError(t, err)
		recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, false, 10*time.Second)
		require.NoError(t, err)
		require.Equal(t, recs[0].GasBurned, estimatedReceipt.GasBurned)
		require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.GasFeeCharged)
		require.Len(t, env.getAccountNFTs(targetAgentID), 1)*/
}

func testEstimateGasOffLedger(t *testing.T, env *ChainEnv) {
	// estimate off-ledger request, then send the same request, assert the gas used/fees match
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	env.DepositFunds(10*isc.Million, keyPair)

	estimationReq := isc.NewOffLedgerRequest(
		env.Chain.ChainID,
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		0,
		1*isc.Million,
	).WithAllowance(isc.NewAssets(5000)).
		WithSender(keyPair.GetPublicKey())

	// Test that the API will fail if the FromAddress is missing
	estimatedReceiptFail, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOffledger{
		RequestBytes: cryptolib.EncodeHex(estimationReq.Bytes()),
	}).Execute()
	require.Error(t, err)
	require.Nil(t, estimatedReceiptFail)
	///

	requestHex := cryptolib.EncodeHex(estimationReq.Bytes())

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOffledger{
		FromAddress:  keyPair.Address().String(),
		RequestBytes: requestHex,
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	client := env.Chain.Client(keyPair)
	par := chainclient.PostRequestParams{
		Allowance: isc.NewAssets(5000),
	}
	par.WithGasBudget(1 * isc.Million)

	req, err := client.PostOffLedgerRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		par,
	)
	require.NoError(t, err)
	rec, err := env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, rec.GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, rec.GasFeeCharged, estimatedReceipt.GasFeeCharged)
}
