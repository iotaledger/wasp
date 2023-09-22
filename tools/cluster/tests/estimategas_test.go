package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func testEstimateGasOnLedger(t *testing.T, env *ChainEnv) {
	// estimate on-ledger request, then send the same request, assert the gas used/fees match
	output := transaction.BasicOutputFromPostData(
		tpkg.RandEd25519Address(),
		isc.EmptyContractIdentity(),
		isc.RequestParameters{
			TargetAddress: env.Chain.ChainAddress(),
			Assets:        isc.NewAssetsBaseTokens(1 * isc.Million),
			Metadata: &isc.SendMetadata{
				TargetContract: accounts.Contract.Hname(),
				EntryPoint:     accounts.FuncTransferAllowanceTo.Hname(),
				Params: map[kv.Key][]byte{
					accounts.ParamAgentID: isc.NewAgentID(&iotago.Ed25519Address{}).Bytes(),
				},
				Allowance: isc.NewAssetsBaseTokens(5000),
				GasBudget: 1 * isc.Million,
			},
		},
	)

	outputBytes, err := output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	require.NoError(t, err)

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsApi.EstimateGasOnledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOnledger{
		OutputBytes: iotago.EncodeHex(outputBytes),
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := env.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	par := chainclient.PostRequestParams{
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID: isc.NewAgentID(&iotago.Ed25519Address{}).Bytes(),
		},
		Allowance: isc.NewAssetsBaseTokens(5000),
	}
	par.WithGasBudget(1 * isc.Million)

	tx, err := accountsClient.PostRequest(accounts.FuncTransferAllowanceTo.Name,
		par,
	)
	require.NoError(t, err)
	recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.GasFeeCharged)
}

func testEstimateGasOffLedger(t *testing.T, env *ChainEnv) {
	// estimate off-ledger request, then send the same request, assert the gas used/fees match
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	env.DepositFunds(10*isc.Million, keyPair)

	estimationReq := isc.NewOffLedgerRequest(env.Chain.ChainID,
		accounts.Contract.Hname(),
		accounts.FuncTransferAllowanceTo.Hname(),
		dict.Dict{
			accounts.ParamAgentID: isc.NewAgentID(&iotago.Ed25519Address{}).Bytes(),
		},
		0,
		1*isc.Million,
	).WithAllowance(isc.NewAssetsBaseTokens(5000)).
		WithSender(keyPair.GetPublicKey())

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsApi.EstimateGasOffledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOffledger{
		RequestBytes: iotago.EncodeHex(estimationReq.Bytes()),
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	accountsClient := env.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	par := chainclient.PostRequestParams{
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID: isc.NewAgentID(&iotago.Ed25519Address{}).Bytes(),
		},
		Allowance: isc.NewAssetsBaseTokens(5000),
	}
	par.WithGasBudget(1 * isc.Million)

	req, err := accountsClient.PostOffLedgerRequest(accounts.FuncTransferAllowanceTo.Name,
		par,
	)
	require.NoError(t, err)
	rec, err := env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, rec.GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, rec.GasFeeCharged, estimatedReceipt.GasFeeCharged)
}
